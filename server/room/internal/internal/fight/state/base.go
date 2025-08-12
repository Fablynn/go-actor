package state

import (
	"go-actor/common/pb"
	"go-actor/library/mlog"
	"go-actor/server/room/internal/internal/fight"
)

type BaseState struct{}

func (s *BaseState) Log(curState pb.GameState) {
	mlog.Infof("当前状态机状态:%v", curState)
}

func (s *BaseState) FlushTime(nowMs int64, extra interface{}) {
	game := extra.(*fight.Fight)
	game.FlushExpireTime(nowMs)
}

func (d *BaseState) OnTick(nowMs int64, curState pb.GameState, extra interface{}) pb.GameState {
	game := extra.(*fight.Fight)
	if game.Timeout <= nowMs {
		// 开始游戏
		return game.GetNextState()
	}
	return curState
}
