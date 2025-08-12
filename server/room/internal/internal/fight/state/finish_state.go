package state

import (
	"go-actor/common/pb"
	"go-actor/library/mlog"
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
	game.FlushExpireTime(nowMs)
	mlog.Infof("结束阶段： 回收战场...")
	game.OnExit()
}

func (d *FinishState) OnTick(nowMs int64, curState pb.GameState, extra interface{}) pb.GameState {
	return curState
}

func (d *FinishState) OnExit(nowMs int64, curState pb.GameState, extra interface{}) {}
