package state

import (
	"go-actor/common/pb"
	"go-actor/server/room/internal/internal/fight"
)

/*
	FUNC: // 玩家操作后阶段 触发回合末事件
*/

type AfterState struct {
	BaseState
}

func (d *AfterState) OnEnter(nowMs int64, curState pb.GameState, extra interface{}) {
	game := extra.(*fight.Fight)
	game.FlushExpireTime(nowMs)
	//mlog.Infof("玩家操作后阶段：")
}

func (d *AfterState) OnTick(nowMs int64, curState pb.GameState, extra interface{}) pb.GameState {
	game := extra.(*fight.Fight)
	return game.GetNextState()
}

func (d *AfterState) OnExit(nowMs int64, curState pb.GameState, extra interface{}) {

}
