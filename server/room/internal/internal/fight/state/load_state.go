package state

import (
	"go-actor/common/pb"
	"go-actor/library/mlog"
	"go-actor/server/room/internal/internal/fight"
)

/*
	FUNC: // 加载战斗环境相关 战场效果
*/

type LoadState struct {
	BaseState
}

func (d *LoadState) OnEnter(nowMs int64, curState pb.GameState, extra interface{}) {
	game := extra.(*fight.Fight)
	game.FlushExpireTime(nowMs)
	mlog.Infof("加载天气 战场环境---结束")
}

func (d *LoadState) OnExit(nowMs int64, curState pb.GameState, extra interface{}) {

}
