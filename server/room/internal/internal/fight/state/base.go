package state

import (
	"go-actor/common/pb"
	"go-actor/library/mlog"
)

type BaseState struct{}

func (s *BaseState) Log(curState pb.GameState) {
	mlog.Infof("当前状态机状态:%v", curState)
}
