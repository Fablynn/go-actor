package actor

import (
	"poker_server/common/pb"
	"poker_server/framework/domain"
	"poker_server/library/encode"
	"poker_server/library/mlog"
	"poker_server/library/uerror"
	"poker_server/library/util"
	"reflect"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
)

const (
	INTER_FLAG             = 1 << 5
	HEAD_FLAG              = 1 << 4
	PROTO_FLAG             = 1 << 3
	RSP_FLAG               = 1 << 2
	NOTIFY_FLAG            = 1 << 1
	BYTES_FLAG             = 1 << 0
	CMD_HANDLER            = int32(HEAD_FLAG | PROTO_FLAG | RSP_FLAG)
	NOTIFY_HANDLER         = int32(HEAD_FLAG | PROTO_FLAG | NOTIFY_FLAG)
	BYTES_HANDLER          = int32(BYTES_FLAG)
	HEAD_BYTES_HANDLER     = int32(HEAD_FLAG | BYTES_FLAG)
	INTERFACE_HANDLER      = int32(INTER_FLAG)
	HEAD_INTERFACE_HANDLER = int32(INTER_FLAG | HEAD_FLAG)
)

var (
	rspType   = reflect.TypeOf((*domain.IRspProto)(nil)).Elem()
	bytesType = reflect.TypeOf((*[]byte)(nil)).Elem()
	protoType = reflect.TypeOf((*proto.Message)(nil)).Elem()
	interType = reflect.TypeOf((*interface{})(nil)).Elem()
	headType  = reflect.TypeOf((*pb.Head)(nil))
	pool      = util.PoolSlice[reflect.Value](6)
)

func get(size int) []reflect.Value {
	rets := pool.Get().([]reflect.Value)
	return rets[:size]
}

func put(v []reflect.Value) {
	pool.Put(v)
}

type FuncInfo struct {
	reflect.Method
	incount int
	flag    int32
}

func parseFuncInfo(m reflect.Method) *FuncInfo {
	ins, outs := m.Type.NumIn(), m.Type.NumOut()
	if outs > 1 {
		return nil
	}
	hasHead := util.Ifelse[int](ins > 1 && m.Type.In(1).AssignableTo(headType), 1, 0)
	isBytes, isProto, isInter := 1, 1, 1
	for i := util.Ifelse[int](hasHead < 1, 1, 2); i < ins; i++ {
		if !m.Type.In(i).AssignableTo(bytesType) {
			isBytes = 0
		}
		if !m.Type.In(i).Implements(protoType) {
			isProto = 0
		}
		if m.Type.In(i).Kind() != reflect.Interface || interType != m.Type.In(i) {
			isInter = 0
		}
	}
	flag := hasHead > 0 && isProto > 0 && ins > 3
	isRsp := util.Ifelse[int](flag, 1, 0)
	for i := util.Ifelse[int](flag, 3, ins); i < ins; i++ {
		if !m.Type.In(i).Implements(rspType) {
			isRsp = 0
		}
	}
	isNotify := util.Ifelse[int](hasHead > 0 && isProto > 0 && ins == 3, 1, 0)
	return &FuncInfo{
		Method:  m,
		incount: ins,
		flag:    int32(isInter<<5 | hasHead<<4 | isProto<<3 | isRsp<<2 | isNotify<<1 | isBytes),
	}
}

// 本地直接调用notify函数
func (f *FuncInfo) localProto(rval reflect.Value, head *pb.Head, args ...interface{}) func() {
	return func() {
		in := get(f.incount)
		defer put(in)
		in[0] = rval
		pos := 1
		if f.flag&HEAD_FLAG > 0 {
			in[1] = reflect.ValueOf(head)
			pos++
		}
		for i := pos; i < f.incount; i++ {
			in[i] = reflect.ValueOf(args[i-pos])
		}

		// 调用函数
		rets := f.Func.Call(in)

		// 处理返回值
		result(head, rets, args)
	}
}

// 本地直接调用cmd函数
func (f *FuncInfo) localCmd(rval reflect.Value, head *pb.Head, args ...interface{}) func() {
	ref := atomic.AddInt32(&head.Reference, 1)
	return func() {
		if head.Dst == nil || head.Src == nil {
			mlog.Error(head, "Actor(%s.%s) head参数错误，args:%v", head.ActorName, head.FuncName, args)
			return
		}

		in := get(f.incount)
		defer put(in)
		in[0] = rval
		in[1] = reflect.ValueOf(head)
		for i := 2; i < f.incount; i++ {
			in[i] = reflect.ValueOf(args[i-2])
		}

		// 调用函数
		rets := f.Func.Call(in)

		// 处理返回值
		if atomic.CompareAndSwapInt32(&head.Reference, ref, ref) {
			response(head, rets, in[3:]...)
		} else {
			result(head, rets, args)
		}
	}
}

// rpc跨服务调用notify函数
func (f *FuncInfo) rpcNotify(rval reflect.Value, head *pb.Head, buf []byte) func() {
	return func() {
		in := get(f.incount)
		defer put(in)
		in[0] = rval
		in[1] = reflect.ValueOf(head)
		in[2] = reflect.New(f.Type.In(2).Elem())
		if err := proto.Unmarshal(buf, in[2].Interface().(proto.Message)); err != nil {
			mlog.Error(head, "参数解析失败: %v", err)
			return
		}
		// 调用函数
		rets := f.Func.Call(in)

		// 处理返回值
		result(head, rets, buf)
	}
}

// rpc跨服务调用cmd函数
func (f *FuncInfo) rpcCmd(rval reflect.Value, head *pb.Head, buf []byte) func() {
	ref := atomic.AddInt32(&head.Reference, 1)
	return func() {
		if head.Dst == nil || head.Src == nil {
			mlog.Error(head, "Actor(%s.%s) head参数错误，buf:%v", head.ActorName, head.FuncName, buf)
			return
		}

		in := get(f.incount)
		defer put(in)
		in[0] = rval
		in[1] = reflect.ValueOf(head)
		for i := 2; i < f.incount; i++ {
			in[i] = reflect.New(f.Type.In(i).Elem())
		}
		if err := proto.Unmarshal(buf, in[2].Interface().(proto.Message)); err != nil {
			mlog.Error(head, "参数解析失败: %v", err)
			return
		}

		// 调用函数
		rets := f.Func.Call(in)

		// 处理返回值
		if atomic.CompareAndSwapInt32(&head.Reference, ref, ref) {
			response(head, rets, in[3:]...)
		} else {
			result(head, rets, buf)
		}
	}
}

// rpc跨服务调用gob函数
func (f *FuncInfo) rpcGob(rval reflect.Value, head *pb.Head, buf []byte) func() {
	return func() {
		pos := 1
		if f.flag&HEAD_FLAG > 0 {
			pos++
		}
		// 解析参数参数
		in, err := encode.Decode(buf, f.Method, pos)
		if err != nil {
			mlog.Error(head, "参数解析失败: %v", err)
			return
		}
		in[0] = rval
		if f.flag&HEAD_FLAG > 0 {
			in[1] = reflect.ValueOf(head)
		}

		// 调用函数
		rets := f.Func.Call(in)

		// 处理返回值
		result(head, rets, buf)
	}
}

func result(head *pb.Head, rets []reflect.Value, buf interface{}) {
	if len(rets) > 0 && !rets[0].IsNil() {
		mlog.Error(head, "error:%v, args:%v", rets[0].Interface(), buf)
	} else {
		mlog.Debug(head, "args:%v", buf)
	}
}

func response(head *pb.Head, results []reflect.Value, rsps ...reflect.Value) {
	var err error
	if len(results) > 0 && !results[0].IsNil() {
		err = results[0].Interface().(error)
	}

	// 处理返回值
	for _, rsp := range rsps {
		if err != nil {
			rspProto := rsp.Interface().(domain.IRspProto)
			rspProto.SetHead(uerror.ToRspHead(err))
		}
		retErr := sendRspFunc(head, rsp.Interface().(proto.Message))
		if err != nil {
			mlog.Error(head, "rsp:%v, error:%v, reterr:%v", rsp.Interface(), err, retErr)
		} else {
			mlog.Debug(head, "rsp:%v, reterr:%v", rsp.Interface(), retErr)
		}
	}
}
