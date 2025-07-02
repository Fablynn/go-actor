package state

import (
	"encoding/json"
	"go-actor/common/card"
	"go-actor/common/config/repository/machine_config"
	"go-actor/common/config/repository/texas_config"
	"go-actor/common/config/repository/texas_test_config"
	"go-actor/common/pb"
	"go-actor/framework"
	"go-actor/library/mlog"
	"go-actor/library/util"
	"go-actor/server/room/internal/texas"
	tutil "go-actor/server/room/internal/texas/util"
)

type FlopState struct{ BaseState }

func (d *FlopState) OnEnter(nowMs int64, curState pb.GameState, extra interface{}) {
	room := extra.(*texas.TexasGame)
	table := room.GetTable()
	machine := room.GetMachine()
	record := room.GetRecord()
	table.CurState = curState
	texasCfg := texas_config.MGetID(room.GameId)
	machineCfg := machine_config.MGetGameType(texasCfg.GameType)

	defer func() {
		buf, _ := json.Marshal(room.TexasRoomData)
		mlog.Debugf("roomId:%d,round:%d,Flop OnEnter: %s", room.RoomId, room.Table.Round, string(buf))
	}()

	// 发送3张公共牌
	pos := 0
	room.Deal(3, func(cursor uint32, cardVal uint32) {
		if texasCfg.IsTest {
			if testCfg := texas_test_config.MGetRound(texas_test_config.MGetRoundKey(room.Table.Round)); testCfg != nil {
				strCard := util.Index[string](testCfg.Publics, pos, "")
				cardVal = util.Index[uint32](card.StrToCard(strCard), 0, cardVal)
			}
			pos++
		}
		record.DealRecord.List = append(record.DealRecord.List, &pb.TexasGameDealRecordInfo{
			DealType: pb.DealType_FLOP,
			Card:     cardVal,
			Cursor:   cursor,
		})
		table.GameData.PublicCardList = append(table.GameData.PublicCardList, cardVal)
	})

	// 设置玩家状态
	dealer := room.GetDealer()
	room.Walk(int(dealer.GameInfo.Position), func(usr *pb.TexasPlayerData) bool {
		switch usr.GameInfo.Operate {
		case pb.OperateType_ALL_IN, pb.OperateType_FOLD:
		default:
			usr.GameInfo.IsChange = false
			usr.GameInfo.GameState = curState
			usr.GameInfo.Operate = pb.OperateType_OperateNone
			usr.GameInfo.BetChips = 0
			usr.GameInfo.PreOperate = pb.OperateType_OperateNone
			usr.GameInfo.PreBetChips = 0
		}
		return true
	})

	// 判断是否全部比牌
	event := &pb.TexasDealEventNotify{
		RoomId:    room.RoomId,
		GameState: int32(curState),
		HandsCard: table.GameData.PublicCardList,
		PotPool:   table.GameData.PotPool,
		Duration:  tutil.GetCurStateTTL(machineCfg, curState) + machine.GetCurStateStartTime() - nowMs,
	}
	table.GameData.MaxBetChips = 0
	cursor := room.GetNext(int(dealer.GameInfo.Position), curState, pb.OperateType_FOLD, pb.OperateType_ALL_IN)
	if cursor != nil {
		// 更新游标
		table.GameData.UidCursor = cursor.GameInfo.Position
		table.GameData.MinRaise = texasCfg.BigBlind
		next := room.GetNext(int(cursor.GameInfo.Position), curState, pb.OperateType_FOLD, pb.OperateType_ALL_IN)
		if next != nil && next.Uid != cursor.Uid {
			event.CurBetChairId = (cursor.ChairId)
		}
	}
	// 发送公共牌
	newHead := &pb.Head{
		Src: framework.NewSrcRouter(pb.RouterType_RouterTypeRoomId, room.RoomId),
		Cmd: uint32(pb.CMD_TEXAS_EVENT_NOTIFY),
	}
	framework.NotifyToClient(room.GetPlayerUidList(), newHead, texas.NewTexasEventNotify(pb.TexasEventType_EVENT_FLOP_CARD, event))
	room.Change()
}
