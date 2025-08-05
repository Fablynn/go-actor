package method

import (
	"go-actor/common/pb"
	"go-actor/framework/define"
	"go-actor/library/encode"
	"go-actor/library/mlog"
	"go-actor/library/uerror"
	"go-actor/library/util"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
)

const (
	HEAD_FLAG           = 1 << 0
	REQ_FLAG            = 1 << 1
	RSP_FLAG            = 1 << 2
	BYTES_FLAG          = 1 << 3
	INTERFACE_FLAG      = 1 << 4
	GOB_FLAG            = 1 << 5
	ERROR_FLAG          = 1 << 6
	HEAD_REQ_RSP_MASK   = HEAD_FLAG | REQ_FLAG | RSP_FLAG
	HEAD_REQ_MASK       = HEAD_FLAG | REQ_FLAG
	HEAD_RSP_MASK       = HEAD_FLAG | RSP_FLAG
	REQ_RSP_MASK        = REQ_FLAG | RSP_FLAG
	HEAD_GOB_RSP_MASK   = HEAD_FLAG | GOB_FLAG | RSP_FLAG
	GOB_RSP_MASK        = GOB_FLAG | RSP_FLAG
	HEAD_BYTES_MASK     = HEAD_FLAG | BYTES_FLAG
	HEAD_INTERFACE_MASK = HEAD_FLAG | INTERFACE_FLAG
)

var (
	headType      = reflect.TypeOf((*pb.Head)(nil))
	reqType       = reflect.TypeOf((*proto.Message)(nil)).Elem()
	rspType       = reflect.TypeOf((*define.IRspProto)(nil)).Elem()
	bytesType     = reflect.TypeOf((*[]byte)(nil)).Elem()
	errorType     = reflect.TypeOf((*error)(nil)).Elem()
	interfaceType = reflect.TypeOf((*interface{})(nil)).Elem()
	args          = util.ArrayPool[reflect.Value](6)
	sendResponse  func(*pb.Head, proto.Message) error
)

func Init(f func(*pb.Head, proto.Message) error) {
	sendResponse = f
}

func get(size int) []reflect.Value {
	rets := args.Get().([]reflect.Value)
	return rets[:size]
}

func put(rets []reflect.Value) {
	args.Put(rets)
}

type Method struct {
	reflect.Method
	mask uint32
}

func NewMethod(m reflect.Method) *Method {
	outs := m.Type.NumOut()
	if outs > 1 || outs == 1 && !m.Type.Out(0).Implements(errorType) {
		return nil
	}
	mask := uint32(0)
	if outs == 1 {
		mask = mask | ERROR_FLAG
	}
	for i := 1; i < m.Type.NumIn(); i++ {
		if m.Type.In(i).AssignableTo(headType) {
			mask = mask | HEAD_FLAG
			continue
		}
		if m.Type.In(i).Implements(rspType) {
			mask = mask | RSP_FLAG
			continue
		}
		if m.Type.In(i).Implements(reqType) {
			mask = mask | REQ_FLAG
			continue
		}
		if m.Type.In(i).AssignableTo(bytesType) {
			mask = mask | BYTES_FLAG
			continue
		}
		if m.Type.In(i) == interfaceType {
			mask = mask | INTERFACE_FLAG
		} else {
			mask = mask | GOB_FLAG
		}
	}
	return &Method{Method: m, mask: mask}
}

func (r *Method) addReference(head *pb.Head) (uint32, uint32) {
	depth := atomic.AddUint32(&head.Depth, 1)
	switch (r.mask | ERROR_FLAG) ^ ERROR_FLAG {
	case HEAD_REQ_RSP_MASK, REQ_RSP_MASK, HEAD_GOB_RSP_MASK, GOB_RSP_MASK:
		return depth, atomic.AddUint32(&head.Reference, 1)
	}
	return depth, 0
}

func (r *Method) Call(rval reflect.Value, head *pb.Head, args ...interface{}) func() {
	depth, ref := r.addReference(head)
	return func() {
		params := get(r.Type.NumIn())
		defer put(params)
		params[0] = rval
		pos := 1
		if r.mask&HEAD_FLAG == HEAD_FLAG {
			params[pos] = reflect.ValueOf(head)
			pos++
		}
		for i := pos; i < r.Type.NumIn(); i++ {
			if arg := args[i-pos]; arg != nil {
				params[i] = reflect.ValueOf(arg)
			} else {
				params[i] = reflect.Zero(r.Type.In(i))
			}
		}

		nowMs := time.Now().UnixMilli()
		mlog.Trace(head, "ActorCall--->>> Params<%v>", valueToInterface(params[pos:]))
		rets := r.Func.Call(params)
		var err error
		if r.mask&ERROR_FLAG == ERROR_FLAG && !rets[0].IsNil() {
			err = rets[0].Interface().(error)
		}
		r.result(nowMs, depth, ref, head, params[pos:], err)
	}
}

func (r *Method) Rpc(rval reflect.Value, head *pb.Head, buf []byte) func() {
	depth, ref := r.addReference(head)
	return func() {
		params := get(r.Type.NumIn())
		defer put(params)
		params[0] = rval
		pos := 1
		if r.mask&HEAD_FLAG == HEAD_FLAG {
			params[pos] = reflect.ValueOf(head)
			pos++
		}
		switch (r.mask | ERROR_FLAG) ^ ERROR_FLAG {
		case HEAD_REQ_RSP_MASK, REQ_RSP_MASK, HEAD_REQ_MASK, REQ_FLAG, HEAD_RSP_MASK, RSP_FLAG:
			for i := pos; i < r.Type.NumIn(); i++ {
				params[i] = reflect.New(r.Type.In(i).Elem())
			}
			if err := proto.Unmarshal(buf, params[pos].Interface().(proto.Message)); err != nil {
				mlog.Errorf("参数解析失败 %v", err)
				return
			}
		case HEAD_GOB_RSP_MASK, GOB_RSP_MASK:
			ll := r.Type.NumIn()
			params[ll-1] = reflect.New(r.Type.In(ll - 1).Elem())
			if err := encode.Decode(buf, params, r.Method, pos, ll-1); err != nil {
				mlog.Errorf("参数解析失败 %v", err)
				return
			}
		case HEAD_BYTES_MASK, BYTES_FLAG, HEAD_INTERFACE_MASK, INTERFACE_FLAG:
			params[pos] = reflect.ValueOf(buf)
		default:
			if err := encode.Decode(buf, params, r.Method, pos); err != nil {
				mlog.Errorf("参数解析失败 %v", err)
				return
			}
		}

		nowMs := time.Now().UnixMilli()
		mlog.Trace(head, "ActorCall--->>> Params<%v>", valueToInterface(params[pos:]))
		rets := r.Func.Call(params)
		var err error
		if r.mask&ERROR_FLAG == ERROR_FLAG && !rets[0].IsNil() {
			err = rets[0].Interface().(error)
		}
		r.result(nowMs, depth, ref, head, params[pos:], err)
	}
}

func (r *Method) result(nowMs int64, depth, ref uint32, head *pb.Head, params []reflect.Value, err error) {
	endMs := time.Now().UnixMilli()
	switch (r.mask | ERROR_FLAG) ^ ERROR_FLAG {
	case HEAD_REQ_RSP_MASK, REQ_RSP_MASK:
		if atomic.CompareAndSwapUint32(&head.Reference, ref, ref) {
			req := params[0].Interface().(proto.Message)
			rsp := params[1].Interface().(define.IRspProto)
			rsp.SetHead(uerror.ToRspHead(err))
			reterr := sendResponse(head, rsp)
			if err != nil {
				mlog.Error(head, "ActorResult--->>> time(%dms) | Req<%v> | Rsp<%v> | SendError<%v> | Error<%v>", endMs-nowMs, req, rsp, reterr, err)
			} else {
				mlog.Trace(head, "ActorResult--->>> time(%dms) | Req<%v> | Rsp<%v> | SendError<%v>", endMs-nowMs, req, rsp, reterr)
			}
		}
	case HEAD_GOB_RSP_MASK, GOB_RSP_MASK:
		if atomic.CompareAndSwapUint32(&head.Reference, ref, ref) {
			ll := len(params)
			rsp := params[ll-1].Interface().(define.IRspProto)
			args := valueToInterface(params[:ll-1])
			rsp.SetHead(uerror.ToRspHead(err))
			reterr := sendResponse(head, rsp)
			if err != nil {
				mlog.Error(head, "ActorResult--->>> time(%dms) | Args<%v> | Rsp<%v> | SendError<%v> | Error<%v>", endMs-nowMs, args, rsp, reterr, err)
			} else {
				mlog.Trace(head, "ActorResult--->>> time(%dms) | Args<%v> | Rsp<%v> | SendError<%v>", endMs-nowMs, args, rsp, reterr)
			}
		}
	case HEAD_REQ_MASK, REQ_FLAG:
		if atomic.CompareAndSwapUint32(&head.Depth, depth, depth) {
			req := params[0].Interface().(proto.Message)
			if err != nil {
				mlog.Error(head, "ActorResult--->>> time(%dms) | Req<%v> | Error<%v>", endMs-nowMs, req, err)
			} else {
				mlog.Trace(head, "ActorResult--->>> time(%dms) | Req<%v>", endMs-nowMs, req)
			}
		}
	case HEAD_RSP_MASK, RSP_FLAG:
		if atomic.CompareAndSwapUint32(&head.Depth, depth, depth) {
			rsp := params[0].Interface().(define.IRspProto)
			if err != nil {
				mlog.Error(head, "ActorResult--->>> time(%dms) | Rsp<%v> | Error<%v>", endMs-nowMs, rsp, err)
			} else {
				mlog.Trace(head, "ActorResult--->>> time(%dms) | Rsp<%v>", endMs-nowMs, rsp)
			}
		}
	default:
		if atomic.CompareAndSwapUint32(&head.Depth, depth, depth) {
			args := valueToInterface(params)
			if err != nil {
				mlog.Error(head, "ActorResult--->>> time(%dms) | Args<%v> | Error<%v>", endMs-nowMs, args, err)
			} else {
				mlog.Trace(head, "ActorResult--->>> time(%dms) | Args<%v>", endMs-nowMs, args)
			}
		}
	}
}

func valueToInterface(params []reflect.Value) (rets []interface{}) {
	for _, arg := range params {
		if arg.IsValid() {
			rets = append(rets, arg.Interface())
		}
	}
	return
}
