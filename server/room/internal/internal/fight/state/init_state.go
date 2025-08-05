package state

import (
	"go-actor/common/pb"
	"go-actor/server/room/internal/internal/fight"
)

/*
	FUNC: // 加载战斗环境相关 战场效果
*/

type InitState struct {
	BaseState
}

func (d *InitState) OnEnter(nowMs int64, curState pb.GameState, extra interface{}) {
	game := extra.(*fight.Fight)
	game.Create()
}

func (d *InitState) OnTick(nowMs int64, curState pb.GameState, extra interface{}) pb.GameState {
	return curState
}

func (d *InitState) OnExit(nowMs int64, curState pb.GameState, extra interface{}) {

}
