package texas_room

import (
	"go-actor/common/config/repository/texas_config"
	"go-actor/common/pb"
	"go-actor/framework"
	"go-actor/framework/actor"
	"go-actor/library/mlog"
	"go-actor/library/uerror"
	"sort"
	"time"
)

type MatchTexasRoom struct {
	actor.Actor
	datas map[uint64]*pb.TexasRoomData
}

func NewMatchTexasRoom(id uint64, datas ...*pb.TexasRoomData) *MatchTexasRoom {
	ret := &MatchTexasRoom{datas: make(map[uint64]*pb.TexasRoomData)}
	for _, data := range datas {
		ret.datas[data.RoomId] = data
	}
	ret.Actor.Register(ret)
	ret.Actor.SetId(id)
	ret.Actor.Start()
	return ret
}

func (d *MatchTexasRoom) HasRoomReq(head *pb.Head, req *pb.HasRoomReq, rsp *pb.HasRoomRsp) error {
	_, ok := d.datas[req.RoomId]
	rsp.IsExist = ok
	return nil
}

func (d *MatchTexasRoom) Query(head *pb.Head, roomId uint64) error {
	if tmps := d.check(); len(tmps) > 0 {
		if err := d.build(tmps); err != nil {
			return err
		}
	}
	roomData, ok := d.datas[roomId]
	if !ok {
		return uerror.New(1, pb.ErrorCode_TEXAS_ROOM_NOT_FOUND, "房间不存在 head:%v, roomId:%d", head, roomId)
	}
	return framework.SendResponse(head, roomData)
}

func (d *MatchTexasRoom) Update(head *pb.Head, roomData *pb.TexasRoomData) error {
	if head == nil || roomData == nil {
		return nil
	}
	d.datas[roomData.RoomId] = roomData
	notify := &pb.UpdateTexasRoomDataNotify{List: []*pb.TexasRoomData{roomData}}
	return actor.SendMsg(&pb.Head{ActorName: "MatchTexasRoomMgr", FuncName: "Collect"}, notify)
}

func (d *MatchTexasRoom) MatchRoomReq(head *pb.Head, req *pb.TexasMatchRoomReq, rsp *pb.TexasMatchRoomRsp) error {
	item, ok := d.datas[req.RoomId]
	if !ok {
		return uerror.NEW(pb.ErrorCode_TEXAS_ROOM_NOT_FOUND, head, "换桌房间不存在 req:%v", req)
	}
	cfg := texas_config.MGetID(item.GameId)
	if cfg == nil {
		return uerror.NEW(pb.ErrorCode_CONFIG_NOT_FOUND, head, "房间配置不存在: %v", item)
	}
	if tmps := d.check(req.RoomId); len(tmps) > 0 {
		if err := d.build(tmps); err != nil {
			return err
		}
	}
	arrs := []*pb.TexasRoomData{}
	for _, val := range d.datas {
		if val.GameId != item.GameId || val.RoomId == req.RoomId {
			continue
		}
		arrs = append(arrs, val)
	}
	sort.Slice(arrs, func(i, j int) bool {
		return arrs[i].OnlineNumber > arrs[j].OnlineNumber
	})
	rsp.RoomId = arrs[0].RoomId
	return nil
}

func (d *MatchTexasRoom) RoomListReq(head *pb.Head, req *pb.TexasRoomListReq, rsp *pb.TexasRoomListRsp) error {
	if tmps := d.check(); len(tmps) > 0 {
		if err := d.build(tmps); err != nil {
			return err
		}
	}
	tmps := map[int32]int64{}
	arrs := map[int32][]*pb.TexasRoomInfo{}
	for _, item := range d.datas {
		cfg := texas_config.MGetID(item.GameId)
		elem := &pb.TexasRoomInfo{
			RoomId:         item.RoomId,
			GameType:       int32(cfg.GameType),
			RoomStage:      int32(cfg.RoomType),
			CoinType:       int32(cfg.CoinType),
			RoomState:      int32(item.RoomState),
			CreateTime:     item.CreateTime,
			FinishTime:     item.CreateTime + cfg.RoomKeepLive*60,
			MinBuyIn:       cfg.MinBuyIn,
			MaxBuyIn:       cfg.MaxBuyIn,
			SmallBlind:     int64(cfg.SmallBlind),
			BigBlind:       int64(cfg.BigBlind),
			MaxPlayerCount: int32(cfg.MaxPlayerCount),
			PlayerCount:    int32(item.OnlineNumber),
		}
		tmps[cfg.ID] += item.OnlineNumber
		arrs[cfg.ID] = append(arrs[cfg.ID], elem)
	}
	for iid, vals := range arrs {
		sort.Slice(vals, func(i, j int) bool {
			if vals[i].PlayerCount == vals[j].PlayerCount {
				return vals[i].RoomId < vals[j].RoomId
			}
			return vals[i].PlayerCount > vals[j].PlayerCount
		})
		for _, item := range vals {
			if item.PlayerCount >= item.MaxPlayerCount {
				continue
			}
			item.PlayerCount = int32(tmps[iid])
			rsp.RoomList = append(rsp.RoomList, item)
			break
		}
	}
	sort.Slice(rsp.RoomList, func(i, j int) bool {
		return rsp.RoomList[i].RoomStage < rsp.RoomList[j].RoomStage
	})
	return nil
}

// 检查是否需要创建新房间
func (d *MatchTexasRoom) check(roomIds ...uint64) (rets []int32) {
	nobuilds := map[int32]struct{}{}
	now := time.Now().Unix()

	filter := map[uint64]struct{}{}
	for _, roomId := range roomIds {
		filter[roomId] = struct{}{}
	}

	for roomId, item := range d.datas {
		if _, ok := filter[roomId]; ok {
			continue
		}

		cfg := texas_config.MGetID(item.GameId)
		if cfg == nil || item.CreateTime+int64(cfg.RoomKeepLive*60) <= now {
			delete(d.datas, roomId)
			mlog.Errorf("德州扑克游戏配置不存在: %d", item.RoomId)
			continue
		}
		if item.Table == nil {
			item.Table = &pb.TexasTableData{CurState: pb.GameState_TEXAS_INIT}
		}
		if item.Table.Players == nil {
			item.Table.Players = make(map[uint64]*pb.TexasPlayerData)
		}
		if item.Table.ChairInfo == nil {
			item.Table.ChairInfo = make(map[uint32]uint64)
		}
		if int32(len(item.Table.ChairInfo)) < int32(cfg.MaxPlayerCount) {
			nobuilds[cfg.ID] = struct{}{}
		}
	}
	id := d.GetId()
	gameType, coinType := pb.GameType(id>>16)&0xFFFF, pb.CoinType(id&0xFFFF)
	for _, cfg := range texas_config.GGetGameTypeMatchTypeCoinType(gameType, pb.MatchType_MatchTypeNone, coinType) {
		if _, ok := nobuilds[cfg.ID]; !ok {
			nobuilds[cfg.ID] = struct{}{}
			rets = append(rets, cfg.ID)
		}
	}
	return
}

// 创建房间
func (d *MatchTexasRoom) build(tmps []int32) error {
	ll := len(tmps)
	id := d.Actor.GetId()
	dst := framework.NewBuilderRouter(uint64(pb.GeneratorType_GeneratorTypeTexas), "BuilderTexasGenerator", "GenRoomIdReq")
	head := framework.NewHead(dst, 0, pb.RouterType_RouterTypeGameType, id)
	newreq := &pb.GenRoomIdReq{
		GeneratorType: pb.GeneratorType_GeneratorTypeTexas,
		GameType:      pb.GameType((id >> 16) & 0xFFFF),
		CoinType:      pb.CoinType(uint32(id & 0xFFFF)),
		Count:         int32(ll),
	}
	newrsp := &pb.GenRoomIdRsp{}
	// 获取房间ID
	if err := framework.Request(head, newreq, newrsp); err != nil {
		return err
	}
	if newrsp.Head != nil {
		return uerror.ToError(newrsp.Head)
	}

	// 生成房间
	notify := &pb.UpdateTexasRoomDataNotify{}
	for i, gameId := range tmps {
		item := &pb.TexasRoomData{
			RoomId:     newrsp.RoomIdList[i],
			GameId:     gameId,
			RoomState:  pb.RoomStatus_RoomStatusWait,
			CreateTime: time.Now().Unix(),
			Table: &pb.TexasTableData{
				CurState:  pb.GameState_TEXAS_INIT,
				Players:   make(map[uint64]*pb.TexasPlayerData),
				ChairInfo: make(map[uint32]uint64),
			},
		}
		d.datas[item.RoomId] = item
		notify.List = append(notify.List, item)
	}
	actor.SendMsg(&pb.Head{ActorName: "MatchTexasRoomMgr", FuncName: "Collect"}, notify)
	return nil
}
