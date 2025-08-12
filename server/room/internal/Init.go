package internal

import (
	"go-actor/common/pb"
	"go-actor/server/room/internal/internal/fight/state"
	"go-actor/server/room/internal/manager"
	"go-actor/server/room/internal/module/machine"
)

var (
	fightMgr = manager.NewFightMgr()
)

func Close() {
	fightMgr.Stop()
}

func Load() {
	fightMgr.Load()
}

func init() {
	machine.RegisterState(pb.GameState_GAME_INIT, &state.InitState{})
	machine.RegisterState(pb.GameState_GAME_LOAD, &state.LoadState{})
	machine.RegisterState(pb.GameState_GAME_PRE, &state.PreState{})
	machine.RegisterState(pb.GameState_GAME_ACT, &state.ActState{})
	machine.RegisterState(pb.GameState_GAME_AFTER, &state.AfterState{})
	machine.RegisterState(pb.GameState_GAME_ENEMY, &state.EnemyState{})
	machine.RegisterState(pb.GameState_GAME_SETTLE, &state.SettleState{})
	machine.RegisterState(pb.GameState_GAME_Finish, &state.FinishState{})
}
