package internal

import (
	"go-actor/common/pb"
	"go-actor/framework/actor"
	"go-actor/library/async"
	"go-actor/library/mlog"
	"go-actor/library/uerror"
	"go-actor/server/game/internal/player"
	"reflect"
	"strconv"
)

var (
	playerMgr = NewPlayerMgr()
)

type PlayerMgr struct {
	actor.Actor
	mgr *actor.ActorMgr
}

func Init() error {
	return nil
}

func GetPlayerMgr() *PlayerMgr {
	return playerMgr
}

func NewPlayerMgr() *PlayerMgr {
	mgr := new(actor.ActorMgr)
	pp := &player.Player{}
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

func (p *PlayerMgr) Kick(uid uint64) {
	act := p.mgr.GetActor(uid)
	if act == nil {
		return
	}
	p.mgr.DelActor(uid)

	p.Add(1)
	async.SafeGo(mlog.Errorf, func() {
		act.Stop()
		p.Done()
	})
}

// 登录请求
func (p *PlayerMgr) Login(head *pb.Head, req *pb.GateLoginRequest, rsp *pb.GateLoginResponse) error {
	mlog.Debugf("PlayerDataPool.Login: rsp:%v", req.PlayerData)
	if act := p.mgr.GetActor(head.Uid); act != nil {
		head.FuncName = "Relogin"
		return act.SendMsg(head, req, rsp)
	}

	// todo add dbsvr
	req.PlayerData = &pb.PlayerData{
		Uid: head.Uid,
		Bag: &pb.PlayerDataBag{
			Items: make(map[uint32]*pb.PbItem),
		},
		Base: &pb.PlayerDataBase{
			PlayerInfo: &pb.PlayerInfo{
				Uid:      head.Uid,
				NickName: "player" + strconv.Itoa(int(head.Uid)),
				Avatar:   "AvatarDefault",
			},
		},
	}

	req.PlayerData.Bag.Items[uint32(pb.CoinType_CoinTypeCoin)] = &pb.PbItem{
		PropId: uint32(pb.CoinType_CoinTypeCoin),
		Count:  100000,
	}

	usr := player.NewPlayer(head.Uid, req.PlayerData)
	usr.Start()
	p.mgr.AddActor(usr)
	return usr.SendMsg(head, req, rsp)
}

// 加入德州请求
func (p *PlayerMgr) TexasJoinRoomReq(head *pb.Head, req *pb.TexasJoinRoomReq, rsp *pb.TexasJoinRoomRsp) error {
	if act := p.mgr.GetActor(head.Uid); act != nil {
		return act.SendMsg(head, req, rsp)
	}
	return uerror.NEW(pb.ErrorCode_GAME_PLAYER_NOT_LOGIN, head, "玩家未登录")
}

func (p *PlayerMgr) TexasQuitRoomReq(head *pb.Head, req *pb.TexasQuitRoomReq, rsp *pb.TexasQuitRoomRsp) error {
	if act := p.mgr.GetActor(head.Uid); act != nil {
		return act.SendMsg(head, req, rsp)
	}
	return uerror.NEW(pb.ErrorCode_GAME_PLAYER_NOT_LOGIN, head, "玩家未登录")
}

func (p *PlayerMgr) TexasBuyInReq(head *pb.Head, req *pb.TexasBuyInReq, rsp *pb.TexasBuyInRsp) error {
	if act := p.mgr.GetActor(head.Uid); act != nil {
		return act.SendMsg(head, req, rsp)
	}
	return uerror.NEW(pb.ErrorCode_GAME_PLAYER_NOT_LOGIN, head, "玩家未登录")
}

func (p *PlayerMgr) TexasChangeReq(head *pb.Head, req *pb.TexasChangeRoomReq, rsp *pb.TexasChangeRoomRsp) error {
	if act := p.mgr.GetActor(head.Uid); act != nil {
		return act.SendMsg(head, req, rsp)
	}
	return uerror.NEW(pb.ErrorCode_GAME_PLAYER_NOT_LOGIN, head, "玩家未登录")
}

// QueryPlayerData 重连信息查询
func (p *PlayerMgr) QueryPlayerData(head *pb.Head, req *pb.QueryPlayerDataReq, rsp *pb.QueryPlayerDataRsp) error {
	if act := p.mgr.GetActor(head.GetActorId()); act != nil {
		return act.SendMsg(head, req, rsp)
	}
	mlog.Infof("==========================head:%v", head)
	return uerror.NEW(pb.ErrorCode_GAME_PLAYER_NOT_LOGIN, head, "玩家未登录")
}
