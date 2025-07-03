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

type PreflopState struct{ BaseState }

func (d *PreflopState) OnEnter(nowMs int64, curState pb.GameState, extra interface{}) {
	room := extra.(*texas.TexasGame)
	room.Table.CurState = curState
	record := room.GetRecord()
	machine := room.GetMachine()
	texasCfg := texas_config.MGetID(room.GameId)
	machineCfg := machine_config.MGetGameType(texasCfg.GameType)

	defer func() {
		buf, _ := json.Marshal(room.TexasRoomData)
		mlog.Debugf("roomId:%d,round:%d,Preflop OnEnter: %s", room.RoomId, room.Table.Round, string(buf))
	}()

	// 发送第一轮手牌
	small := room.GetSmall()
	room.Walk(int(small.GameInfo.Position), func(usr *pb.TexasPlayerData) bool {
		usr.GameInfo.IsChange = false
		usr.GameInfo.GameState = curState
		room.Deal(1, func(cursor uint32, cardVal uint32) {
			if texasCfg.IsTest {
				if testCfg := texas_test_config.MGetRound(texas_test_config.MGetRoundKey(room.Table.Round)); testCfg != nil {
					if hands := util.Index[*pb.TexasTestHandCard](testCfg.Hands, int(usr.ChairId-1), nil); hands != nil {
						cardVal = util.Index[uint32](card.StrToCard(hands.Hand1), 0, cardVal)
					}
				}
			}
			record.DealRecord.List = append(record.DealRecord.List, &pb.TexasGameDealRecordInfo{
				DealType: pb.DealType_HAND,
				Uid:      usr.Uid,
				Card:     cardVal,
				Cursor:   cursor,
			})
			usr.GameInfo.HandCardList = append(usr.GameInfo.HandCardList, cardVal)
		})
		return true
	})

	newHead := &pb.Head{
		Src: framework.NewSrcRouter(pb.RouterType_RouterTypeRoomId, room.RoomId),
		Cmd: uint32(pb.CMD_TEXAS_EVENT_NOTIFY),
	}
	ttl := tutil.GetCurStateTTL(machineCfg, curState) + machine.GetCurStateStartTime() - nowMs

	// 发送第二轮手牌 + 发送通知
	cursorPlayer := room.GetCursor()
	room.Walk(int(small.GameInfo.Position), func(usr *pb.TexasPlayerData) bool {
		room.Deal(1, func(cursor uint32, cardVal uint32) {
			if texasCfg.IsTest {
				if testCfg := texas_test_config.MGetRound(texas_test_config.MGetRoundKey(room.Table.Round)); testCfg != nil {
					if hands := util.Index[*pb.TexasTestHandCard](testCfg.Hands, int(usr.ChairId-1), nil); hands != nil {
						cardVal = util.Index[uint32](card.StrToCard(hands.Hand2), 0, cardVal)
					}
				}
			}
			record.DealRecord.List = append(record.DealRecord.List, &pb.TexasGameDealRecordInfo{
				DealType: pb.DealType_HAND,
				Uid:      usr.Uid,
				Card:     cardVal,
				Cursor:   cursor,
			})
			usr.GameInfo.HandCardList = append(usr.GameInfo.HandCardList, cardVal)
			mlog.Infof("Texas deal hard cards notify per user : %v hardcards: %v", usr.Uid, card.CardList(usr.GameInfo.HandCardList))
			// 发送广播
			framework.NotifyToClient([]uint64{usr.Uid}, newHead, texas.NewTexasEventNotify(pb.TexasEventType_EVENT_DEAL, &pb.TexasDealEventNotify{
				RoomId:        room.RoomId,
				GameState:     int32(curState),
				HandsCard:     usr.GameInfo.HandCardList,
				CurBetChairId: cursorPlayer.ChairId,
				PotPool:       room.Table.GameData.PotPool,
				Duration:      ttl,
			}))
		})
		return true
	})
	room.Change()
}
