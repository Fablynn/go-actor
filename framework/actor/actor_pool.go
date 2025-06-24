package actor

import (
	"poker_server/common/pb"
	"poker_server/framework/domain"
	"poker_server/library/async"
	"poker_server/library/mlog"
	"poker_server/library/timer"
	"poker_server/library/uerror"
	"reflect"
	"sync/atomic"
	"time"
)

type ActorPool struct {
	name     string
	id       uint64
	poolSize int
	pool     []*async.Async
	rval     reflect.Value
	funcs    map[string]*FuncInfo
}

func (d *ActorPool) GetId() uint64 {
	return atomic.LoadUint64(&d.id)
}

func (d *ActorPool) SetId(id uint64) {
	atomic.StoreUint64(&d.id, id)
}

func (d *ActorPool) Start() {
	for _, async := range d.pool {
		async.Start()
	}
}

func (d *ActorPool) Stop() {
	atomic.StoreUint64(&d.id, 0)
	for _, async := range d.pool {
		async.Stop()
	}
}

func (d *ActorPool) GetActorName() string {
	return d.name
}

func (d *ActorPool) Register(ac domain.IActor, sizes ...int) {
	if len(sizes) <= 0 {
		panic("ActorPool注册参数错误，必须指定协程池大小")
	}
	d.poolSize = sizes[0]
	d.pool = make([]*async.Async, d.poolSize)
	for i := 0; i < d.poolSize; i++ {
		d.pool[i] = async.NewAsync()
	}
	d.rval = reflect.ValueOf(ac)
	d.name = parseName(d.rval.Elem().Type())
}

func (d *ActorPool) ParseFunc(tt interface{}) {
	switch vv := tt.(type) {
	case map[string]*FuncInfo:
		d.funcs = vv
	case reflect.Type:
		d.funcs = make(map[string]*FuncInfo)
		for i := 0; i < vv.NumMethod(); i++ {
			m := vv.Method(i)
			d.funcs[m.Name] = parseFuncInfo(m)
		}
	default:
		panic("注册参数错误，必须是方法列表或reflect.Type")
	}
}

func (d *ActorPool) SendMsg(h *pb.Head, args ...interface{}) error {
	mm, ok := d.funcs[h.FuncName]
	if !ok {
		return uerror.New(1, pb.ErrorCode_FUNC_NOT_FOUND, "%s.%s未实现", h.ActorName, h.FuncName)
	}
	pos := h.ActorId % uint64(d.poolSize)
	switch mm.flag {
	case CMD_HANDLER:
		d.pool[pos].Push(mm.localCmd(d.rval, h, args...))
	default:
		d.pool[pos].Push(mm.localProto(d.rval, h, args...))
	}
	return nil
}

func (d *ActorPool) Send(h *pb.Head, buf []byte) error {
	mm, ok := d.funcs[h.FuncName]
	if !ok {
		return uerror.New(1, pb.ErrorCode_FUNC_NOT_FOUND, "%s.%s未实现", h.ActorName, h.FuncName)
	}
	pos := h.ActorId % uint64(d.poolSize)
	switch mm.flag {
	case CMD_HANDLER:
		d.pool[pos].Push(mm.rpcCmd(d.rval, h, buf))
	case NOTIFY_HANDLER:
		d.pool[pos].Push(mm.rpcNotify(d.rval, h, buf))
	case BYTES_HANDLER:
		d.pool[pos].Push(mm.localProto(d.rval, h, buf))
	default:
		d.pool[pos].Push(mm.rpcGob(d.rval, h, buf))
	}
	return nil
}

func (d *ActorPool) RegisterTimer(h *pb.Head, ttl time.Duration, times int32) error {
	return timer.Register(&d.id, func() {
		if err := d.SendMsg(h); err != nil {
			mlog.Errorf("定时器发送消息失败: head:%v, error:%v", h, err)
		}
	}, ttl, times)
}
