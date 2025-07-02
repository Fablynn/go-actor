package manager

import (
	"go-actor/common/pb"
	"go-actor/framework"
	"go-actor/framework/actor"
	"go-actor/library/async"
	"go-actor/library/mlog"
	"go-actor/library/uerror"
	"go-actor/server/room/internal/texas"
	"reflect"
)

type TexasGameMgr struct {
	actor.Actor
	mgr *actor.ActorMgr
}

func NewTexasGameMgr() *TexasGameMgr {
	mgr := new(actor.ActorMgr)
	game := &texas.TexasGame{}
	mgr.Register(game)
	mgr.ParseFunc(reflect.TypeOf(game))
	actor.Register(mgr)

	ret := &TexasGameMgr{mgr: mgr}
	ret.Actor.Register(ret)
	ret.Actor.ParseFunc(reflect.TypeOf(ret))
	ret.Actor.SetId(uint64(pb.DataType_DataTypeTexasRoom))
	ret.Actor.Start()
	actor.Register(ret)
	return ret
}

func (d *TexasGameMgr) Stop() {
	d.mgr.Stop()
	d.Actor.Stop()
	mlog.Infof("TexasGameMgr关闭成功")
}

func (d *TexasGameMgr) Remove(id uint64) {
	if act := d.mgr.GetActor(id); act != nil {
		d.mgr.DelActor(id)
		d.Add(1)
		async.SafeGo(mlog.Errorf, func() {
			act.Stop()
			d.Done()
		})
	}
}

// 加入房间请求
func (d *TexasGameMgr) JoinRoomReq(head *pb.Head, req *pb.TexasJoinRoomReq, rsp *pb.TexasJoinRoomRsp) error {
	_, gameType, coinType := pb.MatchType((req.RoomId>>40)&0xFF), pb.GameType((req.RoomId>>32)&0xFF), pb.CoinType((req.RoomId>>24)&0xFF)
	if act := d.mgr.GetActor(req.RoomId); act != nil {
		return act.SendMsg(head, req, rsp)
	}
	// 请求房间数据
	dst := framework.NewMatchRouter(uint64(gameType)<<16|uint64(coinType), "MatchTexasRoom", "Query")
	newHead := framework.NewHead(dst, 0, pb.RouterType_RouterTypeDataType, uint64(pb.DataType_DataTypeTexasRoom))
	data := &pb.TexasRoomData{}
	if err := framework.Request(newHead, req.RoomId, data); err != nil {
		return err
	}
	// 创建房间
	rr := texas.NewTexasGame(data)
	d.mgr.AddActor(rr)
	return rr.SendMsg(head, req, rsp)
}

func (d *TexasGameMgr) QuitRoomReq(head *pb.Head, req *pb.TexasQuitRoomReq, rsp *pb.TexasQuitRoomRsp) error {
	if act := d.mgr.GetActor(req.RoomId); act != nil {
		return act.SendMsg(head, req, rsp)
	}
	return uerror.NEW(pb.ErrorCode_ACTOR_ID_NOT_FOUND, head, "actor不存在")
}

func (d *TexasGameMgr) SitDownReq(head *pb.Head, req *pb.TexasSitDownReq, rsp *pb.TexasSitDownRsp) error {
	if act := d.mgr.GetActor(req.RoomId); act != nil {
		return act.SendMsg(head, req, rsp)
	}
	return uerror.NEW(pb.ErrorCode_ACTOR_ID_NOT_FOUND, head, "actor不存在")
}

func (d *TexasGameMgr) StandUpReq(head *pb.Head, req *pb.TexasStandUpReq, rsp *pb.TexasStandUpRsp) error {
	if act := d.mgr.GetActor(req.RoomId); act != nil {
		return act.SendMsg(head, req, rsp)
	}
	return uerror.NEW(pb.ErrorCode_ACTOR_ID_NOT_FOUND, head, "actor不存在")
}

func (d *TexasGameMgr) BuyInReq(head *pb.Head, req *pb.TexasBuyInReq, rsp *pb.TexasBuyInRsp) error {
	if act := d.mgr.GetActor(req.RoomId); act != nil {
		return act.SendMsg(head, req, rsp)
	}
	return uerror.NEW(pb.ErrorCode_ACTOR_ID_NOT_FOUND, head, "actor不存在")
}

func (d *TexasGameMgr) DoBetReq(head *pb.Head, req *pb.TexasDoBetReq, rsp *pb.TexasDoBetRsp) error {
	if act := d.mgr.GetActor(req.RoomId); act != nil {
		return act.SendMsg(head, req, rsp)
	}
	return uerror.NEW(pb.ErrorCode_ACTOR_ID_NOT_FOUND, head, "actor不存在")
}

func (d *TexasGameMgr) StatisticsReq(head *pb.Head, req *pb.TexasStatisticsReq, rsp *pb.TexasStatisticsRsp) error {
	if act := d.mgr.GetActor(req.RoomId); act != nil {
		return act.SendMsg(head, req, rsp)
	}
	return uerror.NEW(pb.ErrorCode_ACTOR_ID_NOT_FOUND, head, "actor不存在")
}
