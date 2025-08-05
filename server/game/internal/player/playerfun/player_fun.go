package playerfun

import (
	"go-actor/common/pb"
	"go-actor/framework/actor"
	"go-actor/server/game/internal/player/domain"
)

type PlayerFun struct {
	actor.BaseActor
	funs map[pb.PlayerDataType]domain.IPlayerFun
}

func NewPlayerFun() *PlayerFun {
	return &PlayerFun{
		funs: make(map[pb.PlayerDataType]domain.IPlayerFun),
	}
}

func (d *PlayerFun) RegisterIPlayerFun(tt pb.PlayerDataType, ff domain.IPlayerFun) {
	d.funs[tt] = ff
}

func (d *PlayerFun) GetIPlayerFun(tt pb.PlayerDataType) domain.IPlayerFun {
	return d.funs[tt]
}

func (d *PlayerFun) Walk(f func(pb.PlayerDataType, domain.IPlayerFun) bool) {
	for tt, fun := range d.funs {
		if !f(tt, fun) {
			return
		}
	}
}

func (d *PlayerFun) Complete() {
}

func (d *PlayerFun) Finish() {
}

func (d *PlayerFun) GetInfoFunc() *PlayerInfoFun {
	return d.funs[pb.PlayerDataType_PLAYER_DATA_INFO].(*PlayerInfoFun)
}

func (d *PlayerFun) GetRoomFunc() *PlayerRoomFun {
	return d.funs[pb.PlayerDataType_PLAYER_DATA_ROOM].(*PlayerRoomFun)
}

func (d *PlayerFun) GetBagFunc() *PlayerBagFun {
	return d.funs[pb.PlayerDataType_PLAYER_DATA_BAG].(*PlayerBagFun)
}
