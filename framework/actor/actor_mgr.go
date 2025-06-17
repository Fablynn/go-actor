package actor

import (
	"poker_server/common/pb"
	"poker_server/framework/domain"
	"poker_server/library/mlog"
	"poker_server/library/timer"
	"poker_server/library/uerror"
	"reflect"
	"sync"
	"time"
)

type ActorMgr struct {
	id     uint64
	name   string
	mutex  sync.RWMutex
	actors map[uint64]domain.IActor
	funcs  map[string]*FuncInfo
}

func (d *ActorMgr) GetCount() int {
	return len(d.actors)
}

func (d *ActorMgr) GetActor(id uint64) domain.IActor {
	d.mutex.RLock()
	actor, ok := d.actors[id]
	d.mutex.RUnlock()
	if ok {
		return actor
	}
	mlog.Infof("actors :%+v id:[%v] actor[%v] ok[%v] \n", d.actors, id, actor, ok)
	return nil
}

func (d *ActorMgr) DelActor(id uint64) {
	d.mutex.Lock()
	delete(d.actors, id)
	d.mutex.Unlock()
}

func (d *ActorMgr) AddActor(act domain.IActor) {
	act.ParseFunc(d.funcs)
	id := act.GetId()
	d.mutex.Lock()
	d.actors[id] = act
	d.mutex.Unlock()
}

func (d *ActorMgr) GetId() uint64 {
	return d.id
}

func (d *ActorMgr) SetId(id uint64) {
	d.id = id
}

func (d *ActorMgr) Start() {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, act := range d.actors {
		act.Start()
	}
}

func (d *ActorMgr) Stop() {
	d.id = 0
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, act := range d.actors {
		act.Stop()
	}
}

func (d *ActorMgr) GetActorName() string {
	return d.name
}

func (d *ActorMgr) Register(ac domain.IActor, _ ...int) {
	rtype := reflect.TypeOf(ac)
	d.name = parseName(rtype)
	d.actors = make(map[uint64]domain.IActor)
}

func (d *ActorMgr) ParseFunc(rr interface{}) {
	switch vv := rr.(type) {
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

func (d *ActorMgr) SendMsg(h *pb.Head, args ...interface{}) error {
	if _, ok := d.funcs[h.FuncName]; !ok {
		return uerror.New(1, pb.ErrorCode_FUNC_NOT_FOUND, "%s.%s未实现", h.ActorName, h.FuncName)
	}
	switch h.SendType {
	case pb.SendType_POINT:
		if act := d.GetActor(h.ActorId); act != nil {
			return act.SendMsg(h, args...)
		} else {
			return uerror.New(1, pb.ErrorCode_ACTOR_ID_NOT_FOUND, "Actor不存在: %v", h)
		}
	case pb.SendType_BROADCAST:
		d.mutex.RLock()
		for _, act := range d.actors {
			if err := act.SendMsg(h, args...); err != nil {
				mlog.Errorf("ActorMgr.Broadcast err: %v", err)
			}
		}
		d.mutex.RUnlock()
	default:
		return uerror.New(1, pb.ErrorCode_SEND_TYPE_NOT_SUPPORTED, "未知的发送类型: %v", h.SendType)
	}
	return nil
}

func (d *ActorMgr) Send(h *pb.Head, buf []byte) error {
	if _, ok := d.funcs[h.FuncName]; !ok {
		return uerror.New(1, pb.ErrorCode_FUNC_NOT_FOUND, "%s.%s未实现", h.ActorName, h.FuncName)
	}
	switch h.SendType {
	case pb.SendType_POINT:
		if act := d.GetActor(h.ActorId); act != nil {
			return act.Send(h, buf)
		} else {
			return uerror.New(1, pb.ErrorCode_ACTOR_ID_NOT_FOUND, "Actor不存在: %v", h)
		}
	case pb.SendType_BROADCAST:
		d.mutex.RLock()
		for _, act := range d.actors {
			if err := act.Send(h, buf); err != nil {
				mlog.Errorf("ActorMgr.Broadcast err: %v", err)
			}
		}
		d.mutex.RUnlock()
	default:
		return uerror.New(1, pb.ErrorCode_SEND_TYPE_NOT_SUPPORTED, "未知的发送类型: %v", h.SendType)
	}
	return nil
}

func (d *ActorMgr) RegisterTimer(h *pb.Head, ttl time.Duration, times int32) error {
	return timer.Register(&d.id, func() {
		if err := d.SendMsg(h); err != nil {
			mlog.Errorf("定时器发送消息失败: head:%v, error:%v", h, err)
		}
	}, ttl, times)
}
