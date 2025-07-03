package player

import (
	"go-actor/common/pb"
	"go-actor/common/room_util"
	"go-actor/framework"
	"go-actor/framework/actor"
	"go-actor/library/mlog"
	"go-actor/library/uerror"
	"go-actor/server/game/internal/player/domain"
	"go-actor/server/game/internal/player/factory"
	"go-actor/server/game/internal/player/playerfun"
	"go-actor/server/game/module/reward"
	"time"
)

const (
	TTL = 15 * 60
)

type Player struct {
	actor.Actor
	*playerfun.PlayerFun
	loginTime  int64 // 登录时间
	updateTime int64 // 更新时间
}

func NewPlayer(uid uint64, data *pb.PlayerData) *Player {
	ret := &Player{PlayerFun: playerfun.NewPlayerFun()}
	ret.Actor.Register(ret)
	ret.Actor.SetId(uid)
	return ret
}

func (p *Player) Stop() {
	if roomInfo := p.GetBaseFunc().GetRoomInfo(); roomInfo != nil && roomInfo.RoomId > 0 {
		//todo  clear bag package
	}

	p.Save()
	p.Actor.Stop()
	mlog.Infof("Player关闭成功 uid:%d", p.GetId())
}

func (p *Player) OnTick() {
	if err := p.Save(); err != nil {
		mlog.Errorf("保存玩家数据失败: %v", err)
	}

	// 剔除玩家
	uid := p.GetId()
	if p.updateTime+TTL < time.Now().Unix() { //心跳检测
		dst := framework.NewGateRouter(uid, "GatePlayerMgr", "Kick")
		framework.Send(framework.NewHead(dst, uid, pb.RouterType_RouterTypeUid, uid))
		actor.SendMsg(&pb.Head{ActorName: "PlayerMgr", FuncName: "Kick", ActorId: uid}, uid)
	}
}

func (p *Player) Save() error {
	if !p.PlayerFun.IsChange() {
		return nil
	}

	playerData := &pb.PlayerData{Uid: p.GetId()}
	p.PlayerFun.Walk(func(tt pb.PlayerDataType, fun domain.IPlayerFun) bool {
		fun.Save(playerData)
		return true
	})

	// todo 发送db保存数据
	return nil
}

func (p *Player) Relogin(head *pb.Head, req *pb.GateLoginRequest, rsp *pb.GateLoginResponse) error {
	p.loginTime = time.Now().Unix()
	p.updateTime = p.loginTime
	p.GetBaseFunc().UpdatePlayerInfo(req.PlayerData.Base.PlayerInfo)

	playerData := &pb.PlayerData{Uid: p.GetId()}
	p.PlayerFun.Walk(func(tt pb.PlayerDataType, fun domain.IPlayerFun) bool {
		fun.Save(playerData)
		return true
	})
	mlog.Infof("%d重新登录成功 %v", head.Uid, playerData)
	return framework.Send(framework.SwapToGate(head, head.Uid, "Player", "LoginSuccess"), req)
}

// 登录请求
func (p *Player) Login(head *pb.Head, req *pb.GateLoginRequest, rsp *pb.GateLoginResponse) error {
	// 初始化所有模块
	for tt, f := range factory.FUNCS {
		p.PlayerFun.Set(tt, f(p.PlayerFun))
	}

	// 按照顺序加载模块
	for _, tt := range factory.LoadList {
		fun := p.PlayerFun.Get(tt)
		if err := fun.Load(req.PlayerData); err != nil {
			return err
		}
	}

	// 加载完成回调
	for _, tt := range factory.LoadList {
		fun := p.PlayerFun.Get(tt)
		if err := fun.LoadComplate(); err != nil {
			return err
		}
	}
	p.loginTime = time.Now().Unix()
	p.updateTime = p.loginTime
	p.RegisterTimer(&pb.Head{ActorName: "Player", FuncName: "OnTick", ActorId: head.Uid, Uid: head.Uid}, 5*time.Second, -1)
	mlog.Infof("%d登录成功 %v", head.Uid, req.PlayerData)
	return framework.Send(framework.SwapToGate(head, head.Uid, "Player", "LoginSuccess"), rsp)
}

func (p *Player) HeartRequest(head *pb.Head, req *pb.GateHeartRequest, rsp *pb.GateHeartResponse) error {
	now := time.Now().Unix()
	if p.updateTime+TTL <= now {
		dst := framework.NewGateRouter(head.Uid, "GatePlayerMgr", "Kick")
		framework.Send(framework.NewHead(dst, head.Uid, pb.RouterType_RouterTypeUid, head.Uid))

		actor.SendMsg(&pb.Head{ActorName: "PlayerMgr", FuncName: "Kick", ActorId: head.Uid}, head.Uid)
		return uerror.NEW(pb.ErrorCode_TIME_OUT, head, "心跳超时: req:%v", req)
	}

	p.updateTime = now
	rsp.Utc = req.Utc
	rsp.BeginTime = req.BeginTime
	rsp.EndTime = now
	return nil
}

func (p *Player) RewardReq(head *pb.Head, req *pb.RewardReq, rsp *pb.RewardRsp) error {
	return p.GetBagFunc().RewardReq(head, req, rsp)
}

func (p *Player) ConsumeReq(head *pb.Head, req *pb.ConsumeReq, rsp *pb.ConsumeRsp) error {
	return p.GetBagFunc().ConsumeReq(head, req, rsp)
}

func (p *Player) GetBagReq(head *pb.Head, req *pb.GetBagReq, rsp *pb.GetBagRsp) error {
	return p.GetBagFunc().GetBagReq(head, req, rsp)
}

func (p *Player) TexasJoinRoomReq(head *pb.Head, req *pb.TexasJoinRoomReq, rsp *pb.TexasJoinRoomRsp) error {
	if roomId, err := p.GetBaseFunc().GetTexasRealRoomId(head, req.RoomId); err != nil {
		return err
	} else if req.RoomId != roomId {
		req.RoomId = roomId
	}
	if err := p.texasJoinRoomReq(head, req, rsp); err != nil {
		return err
	}
	return nil
}

func (p *Player) texasJoinRoomReq(head *pb.Head, req *pb.TexasJoinRoomReq, rsp *pb.TexasJoinRoomRsp) error {
	// 加入房间
	baseFun := p.GetBaseFunc()
	req.PlayerInfo = baseFun.GetPlayerInfo()
	newHead := &pb.Head{
		Src: head.Src,
		Dst: framework.NewRoomRouter(req.RoomId, "TexasGameMgr", "JoinRoomReq"),
		Uid: head.Uid,
	}
	if err := framework.Request(newHead, req, rsp); err != nil {
		return err
	} else if rsp.Head != nil {
		return uerror.ToError(rsp.Head)
	}

	// 记录信息
	baseFun.TexasJoinRoom(req.RoomId)
	return nil
}

func (p *Player) TexasQuitRoomReq(head *pb.Head, req *pb.TexasQuitRoomReq, rsp *pb.TexasQuitRoomRsp) error {
	if err := p.texasQuitRoomReq(head, req, rsp); err != nil {
		return err
	}
	return nil
}

func (p *Player) texasQuitRoomReq(head *pb.Head, req *pb.TexasQuitRoomReq, rsp *pb.TexasQuitRoomRsp) error {
	// 判断是否为断线重连
	baseFun := p.GetBaseFunc()
	if roomId, err := baseFun.GetTexasRealRoomId(head, req.RoomId); err != nil {
		return err
	} else {
		req.RoomId = roomId
	}

	// 退出房间
	newHead := &pb.Head{
		Dst: framework.NewRoomRouter(req.RoomId, "TexasGameMgr", "QuitRoomReq"),
		Uid: head.Uid,
	}
	if err := framework.Request(newHead, req, rsp); err != nil {
		return err
	} else if rsp.Head != nil {
		return uerror.ToError(rsp.Head)
	}

	// 道具入背包
	p.GetBagFunc().AddProp(uint32(rsp.CoinType), rsp.Chip)

	// 删除记录
	baseFun.TexasQuitRoom()
	return nil
}

func (p *Player) TexasBuyInReq(head *pb.Head, req *pb.TexasBuyInReq, rsp *pb.TexasBuyInRsp) error {
	// 先扣除玩家道具
	conReq := reward.ToConsumeRequest(pb.CoinType(req.CoinType), req.Chip)
	conRsp := &pb.ConsumeRsp{}
	if err := p.GetBagFunc().ConsumeReq(head, conReq, conRsp); err != nil {
		return err
	}
	return framework.Send(framework.SwapToRoom(head, req.RoomId, "TexasGameMgr", "BuyInReq"), req)
}

func (p *Player) TexasChangeReq(head *pb.Head, req *pb.TexasChangeRoomReq, rsp *pb.TexasChangeRoomRsp) error {
	baseFun := p.GetBaseFunc()
	if roomId, err := baseFun.GetTexasRealRoomId(head, req.RoomId); err != nil {
		return err
	} else {
		req.RoomId = roomId
	}

	// 匹配房间
	newHead := &pb.Head{Src: head.Src, Uid: head.Uid}
	matchType, gameType, coinType := room_util.TexasRoomIdTo(req.RoomId)
	newHead.Dst = framework.NewMatchRouter(room_util.ToMatchGameId(matchType, gameType, coinType), "MatchTexasRoom", "MatchRoomReq")
	matchRsp := &pb.TexasMatchRoomRsp{}
	if err := framework.Request(newHead, &pb.TexasMatchRoomReq{RoomId: req.RoomId}, matchRsp); err != nil {
		return err
	} else if matchRsp.Head != nil {
		return uerror.ToError(matchRsp.Head)
	}

	// 退出房间
	newHead.Dst = framework.NewGameRouter(head.Uid, "Player", "TexasQuitRoomReq")
	quitRsp := &pb.TexasQuitRoomRsp{}
	if err := p.texasQuitRoomReq(newHead, &pb.TexasQuitRoomReq{RoomId: req.RoomId}, quitRsp); err != nil {
		return err
	} else if quitRsp.Head != nil {
		return uerror.ToError(quitRsp.Head)
	}

	// 加入房间
	newHead.Dst = framework.NewGameRouter(head.Uid, "Player", "TexasJoinRoomReq")
	joinRsp := &pb.TexasJoinRoomRsp{}
	joinReq := &pb.TexasJoinRoomReq{
		RoomId:     matchRsp.RoomId,
		PlayerInfo: baseFun.GetPlayerInfo(),
		MatchType:  matchType,
		CoinType:   quitRsp.CoinType,
		BuyInChips: quitRsp.Chip,
	}
	mlog.Debugf("PlayerJoinTexas: req:%v", joinReq)
	if err := p.texasJoinRoomReq(newHead, joinReq, joinRsp); err != nil {
		return err
	} else if joinRsp.Head != nil {
		return uerror.ToError(matchRsp.Head)
	}

	rsp.RoomInfo = joinRsp.RoomInfo
	rsp.PlayerInfo = joinRsp.PlayerInfo
	rsp.TableInfo = joinRsp.TableInfo
	rsp.Duration = joinRsp.Duration
	return nil
}
