package state

import (
	"go-actor/common/pb"
	"go-actor/server/room/internal/internal/fight"
)

/*
	FUNC: // 加载战斗环境相关 战场效果
*/

type FinishState struct {
	BaseState
}

func (d *FinishState) OnEnter(nowMs int64, curState pb.GameState, extra interface{}) {
	game := extra.(*fight.Fight)
	game.Create()
}

func (d *FinishState) OnTick(nowMs int64, curState pb.GameState, extra interface{}) pb.GameState {
	return curState
}

func (d *FinishState) OnExit(nowMs int64, curState pb.GameState, extra interface{}) {

}
