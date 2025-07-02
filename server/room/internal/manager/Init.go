package manager

import (
	"go-actor/common/machine"
	"go-actor/common/pb"
	"go-actor/server/room/internal/texas/state"
)

var (
	texasMgr = NewTexasGameMgr()
)

func Init() error {
	return nil
}

func Close() {
	texasMgr.Stop()
}

func init() {
	machine.RegisterState(pb.GameState_TEXAS_INIT, &state.InitState{})
	machine.RegisterState(pb.GameState_TEXAS_START, &state.StartState{})
	machine.RegisterState(pb.GameState_TEXAS_PRE_FLOP, &state.PreflopState{})
	machine.RegisterState(pb.GameState_TEXAS_FLOP_ROUND, &state.FlopState{})
	machine.RegisterState(pb.GameState_TEXAS_TURN_ROUND, &state.TurnState{})
	machine.RegisterState(pb.GameState_TEXAS_RIVER_ROUND, &state.RiverState{})
	machine.RegisterState(pb.GameState_TEXAS_END, &state.EndState{})
}
