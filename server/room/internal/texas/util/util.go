package util

import "go-actor/common/pb"

func GetCurStateTTL(machineCfg *pb.MachineConfig, curState pb.GameState) int64 {
	switch curState {
	case pb.GameState_TEXAS_START:
		if machineCfg != nil {
			return machineCfg.StartDuration * 1000 // 准备时间
		}
		return 0
	case pb.GameState_TEXAS_PRE_FLOP:
		if machineCfg != nil {
			return machineCfg.PreFlopDuration * 1000 // 准备时间
		}
		return 20000
	case pb.GameState_TEXAS_FLOP_ROUND:
		if machineCfg != nil {
			return machineCfg.FlopDuration * 1000 // 准备时间
		}
		return 20000
	case pb.GameState_TEXAS_TURN_ROUND:
		if machineCfg != nil {
			return machineCfg.TurnDuration * 1000 // 准备时间
		}
		return 20000
	case pb.GameState_TEXAS_RIVER_ROUND:
		if machineCfg != nil {
			return machineCfg.RiverDuration * 1000 // 准备时间
		}
		return 20000
	case pb.GameState_TEXAS_END:
		if machineCfg != nil {
			return machineCfg.EndDuration * 1000 // 准备时间
		}
		return 5000
	default:
		return 0
	}
}
