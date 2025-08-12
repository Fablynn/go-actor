package state

import (
	"go-actor/common/pb"
	"go-actor/library/mlog"
	"go-actor/server/room/internal/internal/fight"
)

/*
	FUNC: // 加载战斗环境相关 战场效果
*/

type SettleState struct {
	BaseState
}

func (d *SettleState) OnEnter(nowMs int64, curState pb.GameState, extra interface{}) {
	game := extra.(*fight.Fight)
	game.FlushExpireTime(nowMs)
	mlog.Infof("结算阶段： 金币奖励30 以下内容三选一...")
}

func (d *SettleState) OnTick(nowMs int64, curState pb.GameState, extra interface{}) pb.GameState {
	game := extra.(*fight.Fight)

	if game.Timeout <= nowMs {
		// 开始游戏
		return game.GetNextState()
	}
	return curState
}

func (d *SettleState) OnExit(nowMs int64, curState pb.GameState, extra interface{}) {

}
