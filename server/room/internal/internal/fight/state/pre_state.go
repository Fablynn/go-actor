package state

import (
	"go-actor/common/pb"
	"go-actor/server/room/internal/internal/fight"
)

/*
	FUNC: // 加载战斗环境相关 战场效果
*/

type PreState struct {
	BaseState
}

func (d *PreState) OnEnter(nowMs int64, curState pb.GameState, extra interface{}) {
	game := extra.(*fight.Fight)
	game.Create()
}

func (d *PreState) OnTick(nowMs int64, curState pb.GameState, extra interface{}) pb.GameState {
	return curState
}

func (d *PreState) OnExit(nowMs int64, curState pb.GameState, extra interface{}) {

}
