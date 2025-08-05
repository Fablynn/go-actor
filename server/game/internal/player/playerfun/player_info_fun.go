package playerfun

import (
	"go-actor/common/pb"
	"go-actor/library/uerror"
	"go-actor/server/game/internal/player/domain"
)

type PlayerInfoFun struct {
	*PlayerFun
	createTime int64               // 账号创建时间
	info       *pb.PlayerInfo      // 玩家信息
	login      *pb.PlayerLoginInfo // 玩家登录信息
}

func NewPlayerInfoFun(fun *PlayerFun) domain.IPlayerFun {
	return &PlayerInfoFun{PlayerFun: fun}
}

func (p *PlayerInfoFun) Load(msg *pb.PlayerData) error {
	if msg == nil || msg.Base == nil || msg.Base.PlayerInfo == nil {
		return uerror.New(pb.ErrorCode_PARAM_INVALID, "玩家基础数据为空")
	}
	if msg.Base.LoginInfo == nil {
		msg.Base.LoginInfo = &pb.PlayerLoginInfo{}
	}
	p.info = msg.Base.PlayerInfo
	p.login = msg.Base.LoginInfo
	return nil
}

func (p *PlayerInfoFun) Save(msg *pb.PlayerData) error {
	if msg == nil {
		return uerror.New(pb.ErrorCode_PARAM_INVALID, "玩家数据为空")
	}
	if msg.Base == nil {
		msg.Base = &pb.PlayerDataBase{}
	}
	msg.Base.CreateTime = p.createTime
	msg.Base.PlayerInfo = &pb.PlayerInfo{
		Uid:      p.info.Uid,
		NickName: p.info.NickName,
		Avatar:   p.info.Avatar,
	}
	msg.Base.LoginInfo = &pb.PlayerLoginInfo{
		LastLoginTime:    p.login.LastLoginTime,
		LastLogoutTime:   p.login.LastLogoutTime,
		CurrentLoginTime: p.login.CurrentLoginTime,
	}
	return nil
}

// 更新登录时间
func (p *PlayerInfoFun) UpdateLogin(now int64) {
	p.login.LastLoginTime = p.login.CurrentLoginTime
	p.login.CurrentLoginTime = now
}

// 更新登出时间
func (p *PlayerInfoFun) UpdateLogout(now int64) {
	p.login.LastLogoutTime = now
}

func (d *PlayerInfoFun) SetPlayerInfo(info *pb.PlayerInfo) {
	if info == nil {
		return
	}
	d.info = info
}

func (d *PlayerInfoFun) GetPlayerInfo() *pb.PlayerInfo {
	return d.info
}

// 查询玩家房间信息
func (p *PlayerInfoFun) QueryPlayerData(head *pb.Head, req *pb.QueryPlayerDataReq, rsp *pb.QueryPlayerDataRsp) error {
	msg := &pb.PlayerData{}
	p.Save(msg)
	roomFun := p.GetRoomFunc()
	roomFun.Save(msg)
	rsp.Data = msg.Base
	return nil
}
