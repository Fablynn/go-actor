package playerfun

import (
	"go-actor/common/pb"
	"go-actor/library/uerror"
	"go-actor/server/game/internal/player/domain"
)

const (
	MATCHTTL = int64(15)
)

type PlayerRoomFun struct {
	*PlayerFun
	data *pb.PlayerRoomInfo
}

func NewPlayerRoomFun(fun *PlayerFun) domain.IPlayerFun {
	return &PlayerRoomFun{PlayerFun: fun}
}

func (p *PlayerRoomFun) Load(msg *pb.PlayerData) error {
	if msg == nil || msg.Base == nil {
		return uerror.New(pb.ErrorCode_PARAM_INVALID, "玩家基础数据为空")
	}
	if msg.Base.RoomInfo == nil {
		msg.Base.RoomInfo = &pb.PlayerRoomInfo{}
	}
	p.data = msg.Base.RoomInfo
	return nil
}

func (p *PlayerRoomFun) Save(msg *pb.PlayerData) error {
	if msg == nil {
		return uerror.New(pb.ErrorCode_PARAM_INVALID, "玩家数据为空")
	}
	if msg.Base == nil {
		msg.Base = &pb.PlayerDataBase{}
	}
	msg.Base.RoomInfo = &pb.PlayerRoomInfo{
		GameType: p.data.GameType,
		RoomId:   p.data.RoomId,
		TableId:  p.data.TableId,
	}
	return nil
}

// 获取房间信息
func (p *PlayerRoomFun) Get() *pb.PlayerRoomInfo {
	return p.data
}
