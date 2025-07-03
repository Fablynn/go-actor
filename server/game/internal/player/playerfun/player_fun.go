package playerfun

import (
	"go-actor/common/pb"
	"go-actor/server/game/internal/player/domain"
)

type PlayerFun struct {
	funs     map[pb.PlayerDataType]domain.IPlayerFun
	isChange bool
}

func NewPlayerFun() *PlayerFun {
	return &PlayerFun{
		funs:     make(map[pb.PlayerDataType]domain.IPlayerFun),
		isChange: false,
	}
}

func (d *PlayerFun) Set(tt pb.PlayerDataType, ff domain.IPlayerFun) {
	d.funs[tt] = ff
}

func (d *PlayerFun) Get(tt pb.PlayerDataType) domain.IPlayerFun {
	return d.funs[tt]
}

func (d *PlayerFun) IsChange() bool {
	return d.isChange
}

func (d *PlayerFun) HasSave() {
	d.isChange = false
}

func (d *PlayerFun) Walk(f func(pb.PlayerDataType, domain.IPlayerFun) bool) {
	for tt, fun := range d.funs {
		if !f(tt, fun) {
			return
		}
	}
}

func (d *PlayerFun) Change() {
	d.isChange = true
}

func (d *PlayerFun) LoadComplate() error {
	return nil
}

func (d *PlayerFun) GetBaseFunc() *PlayerBaseFun {
	return d.funs[pb.PlayerDataType_PLAYER_DATA_BASE].(*PlayerBaseFun)
}

func (d *PlayerFun) GetBagFunc() *PlayerBagFun {
	return d.funs[pb.PlayerDataType_PLAYER_DATA_BAG].(*PlayerBagFun)
}
