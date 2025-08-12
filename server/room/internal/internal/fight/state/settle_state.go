package state

import (
	"go-actor/common/pb"
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
}

func (d *SettleState) OnExit(nowMs int64, curState pb.GameState, extra interface{}) {

}
