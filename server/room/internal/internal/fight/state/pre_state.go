package state

import (
	"go-actor/common/pb"
	"go-actor/server/room/internal/internal/fight"
)

/*
	FUNC: // 战前阶段
*/

type PreState struct {
	BaseState
}

func (d *PreState) OnEnter(nowMs int64, curState pb.GameState, extra interface{}) {
	game := extra.(*fight.Fight)
	game.FlushExpireTime(nowMs)
	//mlog.Infof("战前阶段：")
}

func (d *PreState) OnTick(nowMs int64, curState pb.GameState, extra interface{}) pb.GameState {
	game := extra.(*fight.Fight)
	return game.GetNextState()
}

func (d *PreState) OnExit(nowMs int64, curState pb.GameState, extra interface{}) {

}
