package prop

import (
	"go-actor/framework/actor"
	"go-actor/library/mlog"
	"reflect"
)

type PropMgr struct {
	actor.ActorPool
}

func NewPropMgr() *PropMgr {
	ret := &PropMgr{}
	ret.ActorPool.Register(ret, 100)
	ret.ActorPool.ParseFunc(reflect.TypeOf(ret))
	ret.Start()
	actor.Register(ret)
	return ret
}

func (p *PropMgr) Close() {
	p.Stop()
	mlog.Infof("PropMgr关闭成功")
}
