package player

import (
	"go-actor/common/pb"
	"go-actor/framework/actor"
	"go-actor/framework/cluster"
	"go-actor/framework/recycle"
	"go-actor/library/mlog"
	"go-actor/library/uerror"
	"reflect"
)

type PlayerMgr struct {
	actor.Actor
	mgr *actor.ActorMgr
}

func NewPlayerMgr() *PlayerMgr {
	mgr := new(actor.ActorMgr)
	pp := &Player{}
	mgr.Register(pp)
	mgr.ParseFunc(reflect.TypeOf(pp))
	actor.Register(mgr)

	ret := &PlayerMgr{mgr: mgr}
	ret.Actor.Register(ret)
	ret.Actor.ParseFunc(reflect.TypeOf(ret))
	ret.Start()
	actor.Register(ret)
	return ret
}

func (p *PlayerMgr) Close() {
	p.mgr.Stop()
	p.Actor.Stop()
	mlog.Infof("PlayerMgr关闭成功")
}

func (p *PlayerMgr) Kick(head *pb.Head) {
	act := p.mgr.GetActor(head.Uid)
	if act == nil {
		return
	}
	p.mgr.DelActor(head.Uid)
	recycle.Destroy(act.(*Player))
}

// 登录请求
func (p *PlayerMgr) Login(head *pb.Head, req *pb.GateLoginRequest, rsp *pb.GateLoginResponse) error {
	if act := p.mgr.GetActor(head.Uid); act != nil {
		head.FuncName = "Relogin"
		return act.SendMsg(head, req, rsp)
	}

	if req.PlayerData == nil {
		return uerror.New(pb.ErrorCode_NIL_POINTER, "玩家数据为空: %v", req)
	}
	usr := NewPlayer(head.Uid, req.PlayerData)
	p.mgr.AddActor(usr)
	return usr.SendMsg(head, req, rsp)
}

func (p *PlayerMgr) QueryPlayerData(head *pb.Head, req *pb.QueryPlayerDataReq, rsp *pb.QueryPlayerDataRsp) error {
	if act := p.mgr.GetActor(head.ActorId); act != nil {
		return act.SendMsg(head, req, rsp)
	}
	return cluster.SendToDb(head, "RoomInfoMgr.Query", req)
}
