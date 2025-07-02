package state

import (
	"encoding/json"
	"go-actor/common/config/repository/machine_config"
	"go-actor/common/config/repository/texas_config"
	"go-actor/common/pb"
	"go-actor/framework"
	"go-actor/library/mlog"
	"go-actor/server/room/internal/texas"
	"go-actor/server/room/internal/texas/util"
)

const (
	HOLD_TIME = 2000
)

type BaseState struct{}

func (d *BaseState) OnTick(nowMs int64, curState pb.GameState, extra interface{}) pb.GameState {
	room := extra.(*texas.TexasGame)
	room.Table.CurState = curState
	table := room.GetTable()
	record := room.GetRecord()
	machine := room.GetMachine()
	texasCfg := texas_config.MGetID(room.GameId)
	machineCfg := machine_config.MGetGameType(texasCfg.GameType)

	// 是否直接比牌
	if table.GameData.IsCompare {
		if nowMs-machine.GetCurStateStartTime() < HOLD_TIME {
			return curState
		}
		return room.GetNextState()
	}

	// 玩家超时处理
	usr := room.GetCursor()
	aprev := room.GetPrev(int(usr.GameInfo.Position), curState, pb.OperateType_FOLD)
	next := room.GetNext(int(usr.GameInfo.Position), curState, pb.OperateType_FOLD, pb.OperateType_ALL_IN)
	if util.GetCurStateTTL(machineCfg, curState)+machine.GetCurStateStartTime() <= nowMs {
		usr.PlayerState = pb.PlayerStatus_PlayerStatusQuitGame
		if next == nil {
			if aprev != nil && aprev.GameInfo.BetChips == table.GameData.MaxBetChips && usr.GameInfo.BetChips == table.GameData.MaxBetChips {
				room.Operate(usr, pb.OperateType_CHECK, 0)
			} else {
				room.Operate(usr, pb.OperateType_FOLD, 0)
			}
		} else {
			if aprev != nil && aprev.GameInfo.BetChips == table.GameData.MaxBetChips && usr.GameInfo.BetChips == table.GameData.MaxBetChips && next.GameInfo.BetChips == usr.GameInfo.BetChips {
				room.Operate(usr, pb.OperateType_CHECK, 0)
			} else {
				room.Operate(usr, pb.OperateType_FOLD, 0)
			}
		}
	}

	// 是否有操作
	if !usr.GameInfo.IsChange {
		return curState
	}

	defer func() {
		buf, _ := json.Marshal(room.TexasRoomData)
		mlog.Debugf("roomId:%d,round:%d,%s OnTick: %s", room.RoomId, room.Table.Round, curState.String(), string(buf))
	}()

	// 添加操作记录
	record.OperateRecord.List = append(record.OperateRecord.List, &pb.TexasGameOperateRecordInfo{
		GameState:        curState,
		Uid:              usr.Uid,
		Operate:          usr.GameInfo.Operate,
		BetChips:         usr.GameInfo.BetChips,
		Chips:            usr.Chips,
		TotalPotBetChips: room.Table.GameData.PotPool.TotalBetChips,
	})

	// 先处理玩家操作，进行状态转移
	usr.GameInfo.IsChange = false
	machine.SetCurStateStartTime(nowMs)
	event := &pb.TexasBetEventNotify{
		RoomId:      (room.RoomId),
		ChairId:     (usr.ChairId),
		Chips:       (usr.Chips),
		OperateType: (int32(usr.GameInfo.Operate)),
		BetChips:    (usr.GameInfo.BetChips),
		PotPool:     table.GameData.PotPool,
		MinRaise:    table.GameData.MinRaise,
		MaxBetChips: table.GameData.MaxBetChips,
		Duration:    (machine.GetCurStateStartTime() + util.GetCurStateTTL(machineCfg, curState) - nowMs),
	}

	defer func() {
		room.Change()
		uids := room.GetPlayerUidList()
		newHead := &pb.Head{
			Src: framework.NewSrcRouter(pb.RouterType_RouterTypeRoomId, room.RoomId),
			Cmd: uint32(pb.CMD_TEXAS_EVENT_NOTIFY),
		}
		framework.NotifyToClient(uids, newHead, texas.NewTexasEventNotify(pb.TexasEventType_EVENT_BET, event))
	}()

	// 状态转移
	bigPlayer := room.GetBig()
	anext := room.GetNext(int(usr.GameInfo.Position), curState, pb.OperateType_FOLD)
	switch usr.GameInfo.Operate {
	case pb.OperateType_FOLD:
		// 无任何活跃玩家
		if next == nil {
			if aprev == nil || aprev.Uid == anext.Uid {
				return pb.GameState_TEXAS_END
			} else {
				table.GameData.IsCompare = true
				return room.GetNextState()
			}
		}
		// 只剩一个all in玩家 或者 没有任何all in玩家
		if aprev == nil || aprev.Uid == next.Uid {
			return pb.GameState_TEXAS_END
		}
		// 筹码追平，结束当前论下注
		if aprev.GameInfo.BetChips == usr.GameInfo.BetChips && usr.GameInfo.BetChips == next.GameInfo.BetChips && next.GameInfo.Operate != pb.OperateType_OperateNone {
			return room.GetNextState()
		}
		// 下一个玩家发话
		table.GameData.UidCursor = next.GameInfo.Position
		event.NextChairId = next.ChairId

	case pb.OperateType_ALL_IN:
		// 无任何活跃玩家
		if next == nil {
			if aprev == nil {
				return pb.GameState_TEXAS_END
			} else {
				table.GameData.IsCompare = true
				return room.GetNextState()
			}
		}
		// 下一个玩家发话
		table.GameData.UidCursor = next.GameInfo.Position
		event.NextChairId = next.ChairId
	default:
		// 无任何活跃玩家
		if next == nil {
			if aprev == nil {
				return pb.GameState_TEXAS_END
			} else {
				table.GameData.IsCompare = true
				return room.GetNextState()
			}
		}

		bigBlind := int64(20)
		if texasCfg != nil {
			bigBlind = texasCfg.BigBlind
		}

		// 结束当前轮下注
		prev := room.GetPrev(int(usr.GameInfo.Position), curState, pb.OperateType_FOLD, pb.OperateType_ALL_IN)
		if !(table.CurState == pb.GameState_TEXAS_PRE_FLOP && bigBlind == usr.GameInfo.BetChips && next.Uid == bigPlayer.Uid) &&
			prev.GameInfo.BetChips == usr.GameInfo.BetChips && next.GameInfo.BetChips == usr.GameInfo.BetChips && next.GameInfo.Operate != pb.OperateType_OperateNone {
			return room.GetNextState()
		}
		// 继续执行当前轮
		table.GameData.UidCursor = next.GameInfo.Position
		event.NextChairId = next.ChairId
	}
	return curState
}

func (d *BaseState) OnExit(nowMs int64, curState pb.GameState, extra interface{}) {
	room := extra.(*texas.TexasGame)
	room.UpdateSide(room.GetPlayersByGameState(curState)...)

	defer func() {
		buf, _ := json.Marshal(room.TexasRoomData)
		mlog.Debugf("roomId:%d,round:%d,%s OnTick: %s", room.RoomId, room.Table.Round, curState.String(), string(buf))
	}()
}
