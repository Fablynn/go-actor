package player

import (
	"go-actor/common/pb"
	"go-actor/framework"
	"go-actor/framework/actor"
	"go-actor/framework/cluster"
	"go-actor/library/mlog"
	"go-actor/library/uerror"
	"go-actor/server/game/internal/player/domain"
	"go-actor/server/game/internal/player/factory"
	"go-actor/server/game/internal/player/playerfun"
	"time"
)

const (
	PLAYER_TTL = 15 * 60
)

type Player struct {
	*playerfun.PlayerFun
	heartTime int64 // 调试时间
}

func NewPlayer(uid uint64, data *pb.PlayerData) *Player {
	ret := &Player{PlayerFun: playerfun.NewPlayerFun()}
	ret.Actor.Register(ret)
	ret.Actor.SetId(uid)
	ret.Actor.Start()
	return ret
}

func (p *Player) Close() {
	uid := p.GetId()
	p.Save()
	p.Actor.Stop()
	mlog.Infof("Player玩家%d关闭", uid)
}

func (p *Player) Save() {
	if p.IsChange() {
		// 加载数据
		uid := p.GetId()
		newData := &pb.PlayerData{Uid: uid}
		p.Walk(func(_ pb.PlayerDataType, fun domain.IPlayerFun) bool {
			fun.Save(newData)
			return true
		})

		//// 保存数据
		//head := pb.Head{
		//	Uid: uid,
		//	Src: framework.NewSrcRouter(pb.RouterType_UID, uid),
		//}
		//err := cluster.SendToDb(&head, "PlayerDataMgr.Update", &pb.UpdatePlayerDataNotify{
		//	DataType: pb.DataType_DataTypePlayerData,
		//	Data:     newData,
		//})
		//if err != nil {
		//	mlog.Errorf("玩家%d数据保存失败: %v", uid, newData)
		//} else {
		//	p.Flush()
		//}
	}
}

func (p *Player) OnTick() {
	now := time.Now().Unix()
	if now-p.GetUpdateTime() >= 1 {
		p.Save()
	}
	// 剔除玩家
	if p.heartTime+PLAYER_TTL <= now {
		p.GetInfoFunc().UpdateLogout(now)
		head := pb.Head{
			ActorName: "PlayerMgr",
			FuncName:  "Kick",
			Uid:       p.GetId(),
			Src:       framework.NewSrcRouter(pb.RouterType_UID, p.GetId()),
		}
		actor.SendMsg(&head)
		cluster.SendToGate(&head, "GatePlayerMgr.Kick")
	}
}

// 登录请求
func (p *Player) Login(head *pb.Head, req *pb.GateLoginRequest, rsp *pb.GateLoginResponse) error {
	// 初始化所有模块
	for tt, f := range factory.FUNCS {
		p.PlayerFun.RegisterIPlayerFun(tt, f(p.PlayerFun))
	}
	// 按照顺序加载模块
	for _, tt := range factory.LoadList {
		fun := p.PlayerFun.GetIPlayerFun(tt)
		if err := fun.Load(req.PlayerData); err != nil {
			return err
		}
	}
	// 加载完成回调
	for _, tt := range factory.LoadList {
		fun := p.PlayerFun.GetIPlayerFun(tt)
		fun.Complete()
	}
	// 设置登录信息
	p.Flush()
	p.GetInfoFunc().UpdateLogin(p.GetUpdateTime())
	p.RegisterTimer(&pb.Head{Uid: head.Uid, FuncName: "OnTick"}, 5*time.Second, -1)
	mlog.Info(head, "%d登录成功 %v", head.Uid, req.PlayerData)
	return cluster.SendToGate(head, "Player.LoginSuccess", rsp)
}

// 重新登录
func (p *Player) Relogin(head *pb.Head, req *pb.GateLoginRequest, rsp *pb.GateLoginResponse) error {
	p.Flush()
	p.GetInfoFunc().UpdateLogin(p.GetUpdateTime())
	newData := &pb.PlayerData{Uid: p.GetId()}
	p.PlayerFun.Walk(func(_ pb.PlayerDataType, fun domain.IPlayerFun) bool {
		fun.Save(newData)
		return true
	})
	mlog.Info(head, "%d重新登录成功 %v", head.Uid, newData)
	return cluster.SendToGate(head, "Player.LoginSuccess", rsp)
}

// 心跳请求
func (p *Player) HeartRequest(head *pb.Head, req *pb.GateHeartRequest, rsp *pb.GateHeartResponse) error {
	now := time.Now().Unix()
	if p.heartTime <= 0 {
		p.heartTime = now
	}
	if p.heartTime+PLAYER_TTL <= now {
		return uerror.New(pb.ErrorCode_TIME_OUT, "心跳超时")
	}
	p.heartTime = now
	rsp.Utc = req.Utc
	rsp.BeginTime = req.BeginTime
	rsp.EndTime = now
	return nil
}

// 查询玩家数据
func (p *Player) QueryPlayerData(head *pb.Head, req *pb.QueryPlayerDataReq, rsp *pb.QueryPlayerDataRsp) error {
	return p.GetInfoFunc().QueryPlayerData(head, req, rsp)
}

// 发送奖励请求
func (p *Player) RewardReq(head *pb.Head, req *pb.RewardReq, rsp *pb.RewardRsp) error {
	return p.GetBagFunc().RewardReq(head, req, rsp)
}

// 消耗道具请求
func (p *Player) ConsumeReq(head *pb.Head, req *pb.ConsumeReq, rsp *pb.ConsumeRsp) error {
	return p.GetBagFunc().ConsumeReq(head, req, rsp)
}

// 查询背包请求
func (p *Player) GetBagReq(head *pb.Head, req *pb.GetBagReq, rsp *pb.GetBagRsp) error {
	return p.GetBagFunc().GetBagReq(head, req, rsp)
}
