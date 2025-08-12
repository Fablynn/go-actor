package state

import (
	"go-actor/common/pb"
	"go-actor/server/room/internal/internal/fight"
)

/*
	FUNC: // 加载战斗环境相关 战场效果
*/

type AfterState struct {
	BaseState
}

func (d *AfterState) OnEnter(nowMs int64, curState pb.GameState, extra interface{}) {
	game := extra.(*fight.Fight)
	game.FlushExpireTime(nowMs)
}

func (d *AfterState) OnExit(nowMs int64, curState pb.GameState, extra interface{}) {

}
