package manager

import (
	"go-actor/common/config/repository/character"
	"go-actor/common/config/repository/enemys"
	"go-actor/common/pb"
	"go-actor/framework/actor"
	"go-actor/library/mlog"
	"go-actor/library/safe"
	"go-actor/library/uerror"
	"go-actor/server/room/internal/internal/fight"
	"reflect"
)

type FightMgr struct {
	actor.Actor
	mgr      *actor.ActorMgr
	isFinish bool
}

func NewFightMgr() *FightMgr {
	// 预先注册
	mgr := new(actor.ActorMgr)
	room := &fight.Fight{}
	mgr.Register(room)
	mgr.ParseFunc(reflect.TypeOf(room))
	actor.Register(mgr)

	// 创建房间actor 管理器
	ret := &FightMgr{mgr: mgr}
	ret.Actor.Register(ret)
	ret.Actor.ParseFunc(reflect.TypeOf(ret))
	ret.Actor.Start()
	actor.Register(ret)
	return ret
}

func (d *FightMgr) Load() {
	data := &pb.FightData{
		FightId:    1,
		Characters: character.LGet(),
		Ememys:     enemys.LGet(),
	}

	// 创建房间
	rr := fight.NewFight(data)
	if rr == nil {
		mlog.Infof("战斗服异常: %v", data)
		return
	}

	d.mgr.AddActor(rr)
	rr.Init()
}

// Stop 服务停用
func (d *FightMgr) Stop() {
	// 停止所有游戏
	d.mgr.Stop()

	// 停止自己
	d.Actor.Stop()
	mlog.Infof("FightMgr关闭成功")
}

// Remove 战斗场景超时触发
func (d *FightMgr) Remove(roomID uint64) {
	if act := d.mgr.GetActor(roomID); act != nil {
		d.mgr.DelActor(roomID)
		d.Add(1)
		safe.Go(func() {
			act.Stop()
			d.Done()
		})
	}
}

// Shut 平滑关闭所有子对象
func (d *FightMgr) Shut(head *pb.Head, req *pb.FightShutReq, rsp *pb.FightShutRsp) error {
	d.isFinish = true
	head.SendType = pb.SendType_BROADCAST
	return d.mgr.SendMsg(head)
}

// JoinRoomReq 创建战场
func (d *FightMgr) JoinRoomReq(head *pb.Head, req *pb.CreateFightReq, rsp *pb.CreateFightRsp) error {
	if act := d.mgr.GetActor(req.FightId); act != nil {
		return act.SendMsg(head, req, rsp)
	}

	// 创建房间
	rr := fight.NewFight(req.Data)
	if rr == nil {
		return uerror.New(pb.ErrorCode_CONFIG_NOT_FOUND, "战斗服异常: %v", req.Data)
	}

	d.mgr.AddActor(rr)
	rr.Init()
	return rr.SendMsg(head, req, rsp)
}
