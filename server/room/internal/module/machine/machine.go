package machine

import (
	"fmt"
	"go-actor/common/pb"
)

// 游戏业务通用状态机
type Machine struct {
	startTime int64
	curState  pb.GameState
}

func NewMachine(nowMs int64, state pb.GameState) *Machine {
	// 创建状态机
	return &Machine{curState: state, startTime: nowMs}
}

func (m *Machine) GetState() IState {
	return GetState(m.curState)
}

func (m *Machine) GetCurState() pb.GameState {
	return m.curState
}

func (m *Machine) SetCurStateStartTime(nowMs int64) {
	m.startTime = nowMs
}

func (m *Machine) GetCurStateStartTime() int64 {
	return m.startTime
}

func (m *Machine) Handle(nowMs int64, extra interface{}) {
	// 获取状态机
	handle := GetState(m.curState)
	if handle == nil {
		panic(fmt.Sprintf("Machine状态机未注册状态: %d", m.curState))
	}

	// 状态处理
	nextState := handle.OnTick(nowMs, m.curState, extra)
	if nextState != m.curState {
		m.moveTo(nextState, nowMs, extra)
	}
}

// 状态转移
func (m *Machine) moveTo(newState pb.GameState, nowMs int64, extra interface{}) {
	// 执行旧状态退出逻辑
	state := GetState(m.curState)
	if state == nil {
		panic(fmt.Sprintf("状态机未注册状态: %d", newState))
	}
	state.OnExit(nowMs, m.curState, extra)

	// 更新状态时间
	m.curState = newState
	m.startTime = nowMs

	// 执行进入逻辑
	handle := GetState(m.curState)
	if handle == nil {
		panic(fmt.Sprintf("Machine状态机未注册状态: %d", newState))
	}
	handle.OnEnter(nowMs, m.curState, extra)
}
