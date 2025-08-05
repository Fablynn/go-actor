package machine

import (
	"go-actor/common/pb"
)

var (
	stateMgr = make(map[pb.GameState]IState)
)

// 状态接口
type IState interface {
	OnEnter(nowMs int64, state pb.GameState, extra interface{})
	OnExit(nowMs int64, state pb.GameState, extra interface{})
	OnTick(nowMs int64, state pb.GameState, extra interface{}) pb.GameState
}

type IMachine interface {
	Handle(int64, interface{})        // 执行状态机
	GetCurState() pb.GameState        // 获取当前状态
	GetCurStateStartTime() int64      // 获取状态机开始时间
	SetCurStateStartTime(nowMs int64) // 获取状态机开始时间
}

func RegisterState(val pb.GameState, state IState) {
	if _, ok := stateMgr[val]; ok {
		panic("StateMgr: 注册状态失败，状态已存在")
	}
	stateMgr[val] = state
}

func GetState(val pb.GameState) IState {
	return stateMgr[val]
}
