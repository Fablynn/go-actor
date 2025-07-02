package texas

import (
	"encoding/json"
	"go-actor/common/card"
	"go-actor/common/config/repository/machine_config"
	"go-actor/common/config/repository/texas_config"
	"go-actor/common/machine"
	"go-actor/common/pb"
	"go-actor/framework"
	"go-actor/framework/actor"
	"go-actor/library/mlog"
	"go-actor/library/random"
	"go-actor/library/uerror"
	utillib "go-actor/library/util"
	"go-actor/server/room/internal/texas/util"
	"sort"
	"time"

	"github.com/golang/protobuf/proto"
)

type TexasGame struct {
	actor.Actor
	*pb.TexasRoomData
	infos    map[uint64]*pb.PlayerInfo // 玩家信息
	record   *pb.TexasGameReport       // 游戏记录
	machine  *machine.Machine          // 状态机
	isFinish bool                      // 是否结束
	isChange bool                      // 是否有数据变更
}

func NewTexasGame(data *pb.TexasRoomData) *TexasGame {
	if data.Table == nil {
		data.Table = &pb.TexasTableData{
			CurState:  pb.GameState_TEXAS_INIT,
			Players:   make(map[uint64]*pb.TexasPlayerData),
			ChairInfo: make(map[uint32]uint64),
			GameData:  &pb.TexasGameData{},
		}
	}
	if data.Table.Players == nil {
		data.Table.Players = make(map[uint64]*pb.TexasPlayerData)
	}
	if data.Table.ChairInfo == nil {
		data.Table.ChairInfo = make(map[uint32]uint64)
	}
	if data.Table.GameData == nil {
		data.Table.GameData = &pb.TexasGameData{}
	}

	ret := &TexasGame{
		TexasRoomData: data,
		infos:         make(map[uint64]*pb.PlayerInfo),
	}
	ret.Actor.Register(ret)
	ret.Actor.SetId(data.RoomId)
	ret.Actor.Start()

	// 启动定时器
	head := &pb.Head{SendType: pb.SendType_POINT, ActorName: "TexasGame", FuncName: "OnTick"}
	err := ret.RegisterTimer(head, 50*time.Millisecond, -1)
	if err != nil {
		mlog.Infof("register timer err:%v", err)
	}
	return ret
}

func (d *TexasGame) Stop() {
	mlog.Infof("德州游戏关闭: roomId:%d", d.GetId())
	d.Save()
	d.Actor.Stop()
}

func (d *TexasGame) Finish() {
	d.isFinish = true
}

func (d *TexasGame) Change() {
	d.isChange = true
}

func (d *TexasGame) Save() {
	if !d.isChange {
		return
	}
	cfg := texas_config.MGetID(d.GameId)
	dst := framework.NewMatchRouter(uint64(cfg.GameType&0xFFFF)<<16|uint64(cfg.CoinType&0xFFFF), "MatchTexasRoom", "Update")
	if err := framework.Send(framework.NewHead(dst, 0, pb.RouterType_RouterTypeRoomId, d.RoomId), d.TexasRoomData); err != nil {
		mlog.Errorf("德州扑克房间数据保存失败: %v, error: %v", d.TexasRoomData, err)
	}
	d.isChange = false
}

func (d *TexasGame) OnTick() {
	nowMs := time.Now().UnixMilli()
	if d.machine == nil {
		d.machine = machine.NewMachine(nowMs, pb.GameState_TEXAS_INIT, d)
	} else {
		d.machine.Handle(nowMs, d)
	}

	d.Save()

	if d.isFinish {
		err := actor.SendMsg(&pb.Head{ActorName: "TexasGameMgr", FuncName: "Remove"}, d.GetId())
		if err != nil {
			mlog.Errorf("Texas Game Room Remove Err : %v", err)
		}
	}
}

func (d *TexasGame) GetMachine() *machine.Machine {
	return d.machine
}

func (d *TexasGame) GetRecord() *pb.TexasGameReport {
	return d.record
}

func (d *TexasGame) SetRecord(rr *pb.TexasGameReport) {
	d.record = rr
	if d.record.DealRecord == nil {
		d.record.DealRecord = &pb.TexasGameDealRecord{}
	}
	if d.record.PlayerRecord == nil {
		d.record.PlayerRecord = &pb.TexasGamePlayerRecord{}
	}
	if d.record.OperateRecord == nil {
		d.record.OperateRecord = &pb.TexasGameOperateRecord{}
	}
}

func (d *TexasGame) GetPlayerInfo(uid uint64) *pb.PlayerInfo {
	return d.infos[uid]
}

func (d *TexasGame) GetPlayerUidList() (uids []uint64) {
	for _, usr := range d.Table.Players {
		if usr.PlayerState == pb.PlayerStatus_PlayerStatusQuitRoom {
			continue
		}
		uids = append(uids, usr.Uid)
	}
	return
}

func (d *TexasGame) GetOrNewPlayer(uid uint64) *pb.TexasPlayerData {
	usr := d.Table.Players[uid]
	if usr == nil {
		usr = &pb.TexasPlayerData{Uid: uid, GameInfo: &pb.TexasPlayerGameInfo{}}
		d.Table.Players[uid] = usr
	}
	return usr
}

func (d *TexasGame) GetCursor() *pb.TexasPlayerData {
	gameData := d.Table.GameData
	ll := len(gameData.UidList)
	if ll <= 0 {
		return nil
	}
	return d.Table.Players[gameData.UidList[gameData.UidCursor%uint32(ll)]]
}

func (d *TexasGame) GetDealer() *pb.TexasPlayerData {
	gameData := d.Table.GameData
	return d.Table.Players[d.Table.ChairInfo[gameData.DealerChairId]]
}

func (d *TexasGame) GetSmall() *pb.TexasPlayerData {
	gameData := d.Table.GameData
	return d.Table.Players[d.Table.ChairInfo[gameData.SmallChairId]]
}

func (d *TexasGame) GetBig() *pb.TexasPlayerData {
	gameData := d.Table.GameData
	return d.Table.Players[d.Table.ChairInfo[gameData.BigChairId]]
}

func (d *TexasGame) Walk(pos int, f func(*pb.TexasPlayerData) bool) {
	size := len(d.Table.GameData.UidList)
	for i := pos; i < pos+size; i++ {
		if !f(d.Table.Players[d.Table.GameData.UidList[i%size]]) {
			return
		}
	}
}

func (d *TexasGame) GetPrev(pos int, state pb.GameState, opr ...pb.OperateType) *pb.TexasPlayerData {
	size := len(d.Table.GameData.UidList)
	for i := pos + size - 1; i > pos; i-- {
		flag := false
		usr := d.Table.Players[d.Table.GameData.UidList[i%size]]
		if usr.GameInfo.GameState != state {
			continue
		}
		for _, opr := range opr {
			flag = flag || (usr.GameInfo.Operate == opr)
		}
		if flag {
			continue
		}
		return usr
	}
	return nil
}

func (d *TexasGame) GetNext(pos int, state pb.GameState, oprs ...pb.OperateType) *pb.TexasPlayerData {
	size := len(d.Table.GameData.UidList)
	for i := pos + 1; i < pos+size; i++ {
		flag := false
		usr := d.Table.Players[d.Table.GameData.UidList[i%size]]
		if usr.GameInfo.GameState != state {
			continue
		}
		for _, opr := range oprs {
			flag = flag || (usr.GameInfo.Operate == opr)
		}
		if flag {
			continue
		}
		return usr
	}
	return nil
}

func (d *TexasGame) GetPlayersByGameState(state pb.GameState) (users []*pb.TexasPlayerData) {
	for _, uid := range d.Table.GameData.UidList {
		usr := d.Table.Players[uid]
		if usr.GameInfo.GameState == state {
			users = append(users, usr)
		}
	}
	sort.Slice(users, func(i, j int) bool {
		if users[i].GameInfo.BetChips == users[j].GameInfo.BetChips {
			return users[i].GameInfo.Operate == pb.OperateType_ALL_IN
		}
		return users[i].GameInfo.BetChips < users[j].GameInfo.BetChips
	})
	return
}

func (d *TexasGame) GetGamePlayers() (users []*pb.TexasPlayerData) {
	for _, usr := range d.Table.Players {
		if usr.ChairId > 0 && usr.PlayerState == pb.PlayerStatus_PlayerStatusJoinGame {
			users = append(users, usr)
		}
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].ChairId < users[j].ChairId
	})
	return
}

func (d *TexasGame) Operate(usr *pb.TexasPlayerData, opr pb.OperateType, val int64) {
	switch opr {
	case pb.OperateType_CHECK, pb.OperateType_FOLD:
		usr.GameInfo.Operate = opr
		usr.GameInfo.IsChange = true
	case pb.OperateType_BET_BIG_BLIND, pb.OperateType_BET_SMALL_BLIND:
		usr.GameInfo.Operate = opr
		usr.GameInfo.IsChange = true
		usr.GameInfo.TotalBetChips += val
		usr.GameInfo.BetChips += val
		usr.Chips -= val
		d.Table.GameData.MinRaise = val
		d.Table.GameData.PotPool.TotalBetChips += val
	default:
		usr.GameInfo.Operate = opr
		usr.GameInfo.IsChange = true
		usr.GameInfo.TotalBetChips += val
		usr.GameInfo.BetChips += val
		d.Table.GameData.PotPool.TotalBetChips += val
	}
	if d.Table.GameData.MaxBetChips < usr.GameInfo.BetChips {
		d.Table.GameData.MaxBetChips = usr.GameInfo.BetChips
	}
}

// GetNextState 获取下一个状态
func (d *TexasGame) GetNextState() pb.GameState {
	switch d.machine.GetCurState() {
	case pb.GameState_TEXAS_INIT:
		return pb.GameState_TEXAS_START
	case pb.GameState_TEXAS_START:
		return pb.GameState_TEXAS_PRE_FLOP
	case pb.GameState_TEXAS_PRE_FLOP:
		return pb.GameState_TEXAS_FLOP_ROUND
	case pb.GameState_TEXAS_FLOP_ROUND:
		return pb.GameState_TEXAS_TURN_ROUND
	case pb.GameState_TEXAS_TURN_ROUND:
		return pb.GameState_TEXAS_RIVER_ROUND
	case pb.GameState_TEXAS_RIVER_ROUND:
		return pb.GameState_TEXAS_END
	default:
		return pb.GameState_TEXAS_INIT
	}
}

func (d *TexasGame) Shuffle(times int) {
	gameData := d.Table.GameData
	gameData.CardCursor = 0
	for i := uint32(0); i <= 3; i++ {
		for j := uint32(2); j <= 14; j++ {
			gameData.CardList = append(gameData.CardList, (1<<(16+i))|j)
		}
	}
	for j := 0; j < times; j++ {
		for i := 0; i < len(gameData.CardList); i++ {
			pos := random.Intn(len(gameData.CardList))
			gameData.CardList[i], gameData.CardList[pos] = gameData.CardList[pos], gameData.CardList[i]
		}
	}
}

func (d *TexasGame) Deal(count uint32, f func(uint32, uint32)) {
	table := d.Table
	for i := table.GameData.CardCursor; i < table.GameData.CardCursor+count; i++ {
		if f != nil {
			f(i, table.GameData.CardList[i])
		}
	}
	table.GameData.CardCursor += count
}

func (d *TexasGame) UpdateBest(usr *pb.TexasPlayerData) {
	tmps := make([]uint32, 0, len(d.Table.GameData.PublicCardList)+len(usr.GameInfo.HandCardList))
	tmps = append(tmps, d.Table.GameData.PublicCardList...)
	tmps = append(tmps, usr.GameInfo.HandCardList...)

	cardType, val, list := Compare(tmps...)
	mlog.Debugf("UpdateBest --->>> roomId:%d, round:%d, cardType:%s, bestValue:%d, best:%s", d.RoomId, d.Table.Round, cardType.String(), val, card.CardList(list).String())

	usr.GameInfo.BestCardType = cardType
	usr.GameInfo.BestCardValue = val
	usr.GameInfo.BestCardList = list
}

// UpdateSide 更新边池
func (d *TexasGame) UpdateSide(users ...*pb.TexasPlayerData) {
	count := len(users)
	prevBetChips := int64(0)
	game := d.Table.GameData
	for i, usr := range users {
		switch usr.GameInfo.Operate {
		case pb.OperateType_FOLD:
			game.PotPool.BetChips += usr.GameInfo.BetChips - prevBetChips
		case pb.OperateType_ALL_IN:
			add := int64(count-i) * (usr.GameInfo.BetChips - prevBetChips)
			if add > 0 {
				game.PotPool.PotList = append(game.PotPool.PotList, &pb.TexasPotData{
					PotType: 1,
					Chips:   game.PotPool.BetChips + add,
					UidList: playerToUid(users[i:]...),
				})
				game.PotPool.BetChips = 0
			}
			prevBetChips = usr.GameInfo.BetChips
		default:
			game.PotPool.BetChips += usr.GameInfo.BetChips - prevBetChips
		}
		usr.GameInfo.BetChips = 0
	}
}

func playerToUid(users ...*pb.TexasPlayerData) (rets []uint64) {
	for _, usr := range users {
		rets = append(rets, usr.GetUid())
	}
	return
}

func (d *TexasGame) UpdateMain(users ...*pb.TexasPlayerData) {
	if len(users) > 0 {
		game := d.Table.GameData
		game.PotPool.PotList = append(game.PotPool.PotList, &pb.TexasPotData{
			PotType: 0,
			Chips:   game.PotPool.BetChips,
			UidList: playerToUid(users...),
		})
		game.PotPool.BetChips = 0
	}
}

func (d *TexasGame) getWinners(pot *pb.TexasPotData) (users []*pb.TexasPlayerData) {
	for _, uid := range pot.UidList {
		users = append(users, d.Table.Players[uid])
		usr := d.Table.Players[uid]
		mlog.Debugf("getWinners --->>> roomId:%d, round:%d, cardType:%s, bestValue:%d, best:%s, pot:%v",
			d.RoomId, d.Table.Round, usr.GameInfo.BestCardType.String(), usr.GameInfo.BestCardValue, card.CardList(usr.GameInfo.BestCardList).String(), pot)
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].GameInfo.BestCardValue > users[j].GameInfo.BestCardValue
	})
	j := -1
	maxValue := users[0].GameInfo.BestCardValue
	for _, usr := range users {
		if usr.GameInfo.BestCardValue == maxValue {
			j++
			users[j] = usr
		} else {
			break
		}
	}
	users = users[:j+1]
	return
}

// Reward 计算获胜者筹码
func (d *TexasGame) Reward(wins map[uint64]int64, chips map[uint64]int64, srvs map[uint64]int64) (total int64) {
	rate, maxLimit, isNoFlopNoDrop := int64(0), int64(0), true
	if cfg := texas_config.MGetID(d.GameId); cfg != nil {
		isNoFlopNoDrop = cfg.IsNoFlopNoDrop
		rate = cfg.RakeRate
		ll := len(cfg.RakeUpLimit)
		maxLimit = utillib.Index[int64](cfg.RakeUpLimit, len(d.Table.GameData.UidList)-1, cfg.RakeUpLimit[ll-1])
	}
	for _, uid := range d.Table.GameData.UidList {
		wins[uid] = 0
		chips[uid] -= d.Table.Players[uid].GameInfo.TotalBetChips
	}
	for _, pot := range d.Table.GameData.PotPool.PotList {
		users := d.getWinners(pot)
		lusers := int64(len(users))
		add := pot.Chips/lusers - pot.Chips/int64(len(pot.UidList))
		vsrv := int64(0)
		if !(isNoFlopNoDrop && len(d.Table.GameData.PublicCardList) <= 0) && len(pot.UidList) > 1 {
			diff := pot.Chips * rate / 10000
			incr := maxLimit - total
			if incr >= 0 && incr < diff {
				diff = incr
			}
			if incr < 0 || diff > 0 && diff >= add {
				diff = 0
			}
			total += diff
			pot.Chips -= diff
			vsrv = diff / lusers
		}
		val := pot.Chips / lusers
		for _, usr := range users {
			wins[usr.Uid] += val
			srvs[usr.Uid] += vsrv
		}
	}
	return
}

func (d *TexasGame) StatisticsReq(head *pb.Head, req *pb.TexasStatisticsReq, rsp *pb.TexasStatisticsRsp) {
	buf, _ := proto.Marshal(d.TexasRoomData)
	newData := &pb.TexasRoomData{}
	proto.Unmarshal(buf, newData)
	newData.Table.GameData = nil
	rsp.Data = newData
}

func (d *TexasGame) JoinRoomReq(head *pb.Head, req *pb.TexasJoinRoomReq, rsp *pb.TexasJoinRoomRsp) error {
	texasCfg := texas_config.MGetID(d.GameId)
	if texasCfg == nil {
		return uerror.NEW(pb.ErrorCode_CONFIG_NOT_FOUND, head, "德州扑克配置不存在: %d", d.GameId)
	}
	machineCfg := machine_config.MGetGameType(texasCfg.GameType)
	if machineCfg == nil {
		return uerror.NEW(pb.ErrorCode_CONFIG_NOT_FOUND, head, "德州扑克状态机配置不存在: %d", texasCfg.GameType)
	}
	usr, ok := d.Table.Players[head.Uid]
	if !ok {
		usr = d.GetOrNewPlayer(head.Uid)
		d.Change()
	}
	d.infos[head.Uid] = req.PlayerInfo

	if usr.PlayerState == pb.PlayerStatus_PlayerStatusNone || usr.PlayerState == pb.PlayerStatus_PlayerStatusQuitRoom {
		usr.PlayerState = pb.PlayerStatus_PlayerStatusJoinRoom
		d.OnlineNumber++
		d.Change()
	}
	if req.BuyInChips > 0 {
		d.TotalBuyinChips += req.BuyInChips
		usr.TotalBuyin += req.BuyInChips
		usr.Chips += req.BuyInChips
		d.Change()
	}

	buf, _ := proto.Marshal(d.Table)
	newTable := &pb.TexasTableData{}
	proto.Unmarshal(buf, newTable)
	if newTable.GameData != nil {
		newTable.GameData.CardCursor = 0
		newTable.GameData.CardList = nil
	}

	rsp.Duration = util.GetCurStateTTL(machineCfg, d.Table.CurState)
	rsp.TableInfo = newTable
	rsp.PlayerInfo = d.infos
	rsp.RoomInfo = &pb.TexasRoomInfo{
		RoomId:         d.RoomId,
		GameType:       int32(texasCfg.GameType),
		RoomStage:      int32(texasCfg.RoomType),
		CoinType:       int32(texasCfg.CoinType),
		RoomState:      int32(d.RoomState),
		CreateTime:     d.CreateTime,
		FinishTime:     texasCfg.RoomKeepLive * 60,
		MinBuyIn:       texasCfg.MinBuyIn,
		MaxBuyIn:       texasCfg.MaxBuyIn,
		SmallBlind:     texasCfg.SmallBlind,
		BigBlind:       texasCfg.BigBlind,
		MaxPlayerCount: texasCfg.MaxPlayerCount,
		PlayerCount:    int32(len(d.Table.Players)),
	}
	return nil
}

func (d *TexasGame) QuitRoomReq(head *pb.Head, req *pb.TexasQuitRoomReq, rsp *pb.TexasQuitRoomRsp) error {
	usr, ok := d.Table.Players[head.Uid]
	if !ok || usr == nil {
		return uerror.NEW(pb.ErrorCode_TEXAS_PLAYER_NOT_IN_ROOM, head, "玩家不在房间内: %d", head.Uid)
	}
	if usr.ChairId > 0 {
		if d.Table.CurState != pb.GameState_TEXAS_INIT && d.Table.CurState != pb.GameState_TEXAS_END {
			return uerror.NEW(pb.ErrorCode_TEXAS_PLAYER_IN_GAME, head, "玩家正在游戏中: %d", head.Uid)
		}
		newRsp := &pb.TexasStandUpRsp{}
		d.StandUpReq(&pb.Head{Uid: head.Uid}, &pb.TexasStandUpReq{RoomId: req.RoomId, ChairId: usr.ChairId}, newRsp)
		rsp.ChairId = newRsp.ChairId
		rsp.RoomId = newRsp.RoomId
	}
	cfg := texas_config.MGetID(d.GameId)
	if cfg == nil {
		return uerror.NEW(pb.ErrorCode_CONFIG_NOT_FOUND, head, "游戏配置不存在: GameId:%d", d.GameId)
	}

	rsp.RoomId = req.RoomId
	rsp.CoinType = cfg.CoinType
	rsp.Chip = usr.Chips

	if usr.PlayerState != pb.PlayerStatus_PlayerStatusQuitRoom && usr.PlayerState != pb.PlayerStatus_PlayerStatusNone {
		d.OnlineNumber--
		usr.Chips = 0
		usr.PlayerState = pb.PlayerStatus_PlayerStatusQuitRoom
		delete(d.infos, head.Uid)
		d.Change()
	}
	return nil
}

func (d *TexasGame) SitDownReq(head *pb.Head, req *pb.TexasSitDownReq, rsp *pb.TexasSitDownRsp) error {
	usr, ok := d.Table.Players[head.Uid]
	if !ok || usr == nil {
		return uerror.NEW(pb.ErrorCode_TEXAS_PLAYER_NOT_IN_ROOM, head, "玩家不在房间内: %d", head.Uid)
	}
	if usr.ChairId > 0 {
		return uerror.NEW(pb.ErrorCode_TEXAS_PLAYER_IN_GAME, head, "玩家正在游戏中: %d", head.Uid)
	}
	texasCfg := texas_config.MGetID(d.GameId)
	if texasCfg == nil {
		return uerror.NEW(pb.ErrorCode_CONFIG_NOT_FOUND, head, "德州扑克配置不存在: %d", d.GameId)
	}
	machineCfg := machine_config.MGetGameType(texasCfg.GameType)
	if machineCfg == nil {
		return uerror.NEW(pb.ErrorCode_CONFIG_NOT_FOUND, head, "德州扑克状态机配置不存在: %d", texasCfg.GameType)
	}

	if len(d.Table.ChairInfo) >= int(texasCfg.MaxPlayerCount) {
		return uerror.NEW(pb.ErrorCode_TEXAS_ROOM_IS_FULL, head, "房间已满: %d", d.RoomId)
	}
	if usr.Chips < texasCfg.BigBlind {
		return uerror.NEW(pb.ErrorCode_TEXAS_CHIPS_NOT_ENOUGH, head, "玩家筹码不足: %d", head.Uid)
	}
	// 检查座位是否被占用
	if cur, ok := d.Table.ChairInfo[req.ChairId]; ok {
		if head.Uid == cur {
			return uerror.NEW(pb.ErrorCode_TEXAS_PLAYER_SIT_DOWN_REPEATED, head, "玩家已在座位上: %d", head.Uid)
		}
		return uerror.NEW(pb.ErrorCode_TEXAS_CHAIR_HAS_OCCUPIED, head, "座位已被占用: %d", req.ChairId)
	}

	// 加入牌桌
	usr.PlayerState = pb.PlayerStatus_PlayerStatusJoinGame
	d.Table.ChairInfo[req.ChairId] = head.Uid
	usr.ChairId = req.ChairId
	d.Change()

	// 广播消息
	newHead := &pb.Head{
		Src: framework.NewSrcRouter(pb.RouterType_RouterTypeRoomId, d.RoomId, "TexasGame", "SitDownReq"),
		Cmd: uint32(pb.CMD_TEXAS_EVENT_NOTIFY),
		Uid: head.Uid,
	}
	mlog.Infof("SitDown: %v", usr)
	framework.NotifyToClient(d.GetPlayerUidList(), newHead, NewTexasEventNotify(pb.TexasEventType_EVENT_SIT_DOWN, &pb.TexasPlayerEventNotify{
		RoomId:     d.RoomId,
		ChairId:    req.ChairId,
		Player:     usr,
		Uid:        head.Uid,
		PlayerInfo: d.infos[head.Uid],
	}))
	return nil
}

func NewTexasEventNotify(evenType pb.TexasEventType, event proto.Message) *pb.TexasEventNotify {
	jbuf, _ := json.Marshal(event)
	buf, _ := proto.Marshal(event)
	mlog.Debugf("德州扑克发送通知: %s: %s, buf:%v", evenType.String(), string(jbuf), buf)
	return &pb.TexasEventNotify{
		Event:   evenType,
		Content: buf,
	}
}

// StandUpReq 站起请求
func (d *TexasGame) StandUpReq(head *pb.Head, req *pb.TexasStandUpReq, rsp *pb.TexasStandUpRsp) error {
	usr, ok := d.Table.Players[head.Uid]
	if !ok || usr == nil {
		return uerror.NEW(pb.ErrorCode_TEXAS_PLAYER_NOT_IN_ROOM, head, "玩家不在房间内: %d", head.Uid)
	}
	if usr.ChairId > 0 && usr.ChairId != req.ChairId {
		return uerror.NEW(pb.ErrorCode_PARAM_INVALID, head, "请求参数错误: %v", req)
	}
	if usr.ChairId <= 0 {
		return uerror.NEW(pb.ErrorCode_TEXAS_PLAYER_HAS_STAND_UP, head, "玩家已经站起来: %v", usr)
	}

	d.Change()
	if d.Table.CurState != pb.GameState_TEXAS_INIT && d.Table.CurState != pb.GameState_TEXAS_END {
		usr.PlayerState = pb.PlayerStatus_PlayerStatusQuitGame
		return nil
	}

	usr.ChairId = 0
	delete(d.Table.ChairInfo, req.ChairId)
	usr.PlayerState = pb.PlayerStatus_PlayerStatusQuitGame

	// 广播消息
	newHead := &pb.Head{
		Src: framework.NewSrcRouter(pb.RouterType_RouterTypeRoomId, d.RoomId, "TexasGame", "StandUpReq"),
		Cmd: uint32(pb.CMD_TEXAS_EVENT_NOTIFY),
		Uid: head.Uid,
	}
	framework.NotifyToClient(d.GetPlayerUidList(), newHead, NewTexasEventNotify(pb.TexasEventType_EVENT_STAND_UP, &pb.TexasPlayerEventNotify{
		RoomId:     d.RoomId,
		ChairId:    req.ChairId,
		PlayerInfo: d.infos[head.Uid],
	}))
	return nil
}

// BuyInReq 买入请求
func (d *TexasGame) BuyInReq(head *pb.Head, req *pb.TexasBuyInReq, rsp *pb.TexasBuyInRsp) error {
	usr, ok := d.Table.Players[head.Uid]
	if !ok || usr == nil {
		return uerror.NEW(pb.ErrorCode_TEXAS_PLAYER_NOT_IN_ROOM, head, "玩家不在房间内: %d", head.Uid)
	}
	texasCfg := texas_config.MGetID(d.GameId)
	if texasCfg == nil {
		return uerror.NEW(pb.ErrorCode_CONFIG_NOT_FOUND, head, "德州扑克配置不存在: %d", d.GameId)
	}

	d.TotalBuyinChips += req.Chip
	usr.TotalBuyin += req.Chip
	usr.Chips += req.Chip
	d.Change()

	rsp.RoomId = d.RoomId
	rsp.CoinType = int32(texasCfg.CoinType)
	rsp.Chip = usr.Chips
	return nil
}

// DoBetReq 下注请求
func (d *TexasGame) DoBetReq(head *pb.Head, req *pb.TexasDoBetReq, rsp *pb.TexasDoBetRsp) error {
	usr := d.GetCursor()
	if usr == nil {
		return uerror.NEW(pb.ErrorCode_TEXAS_PLAYER_NOT_IN_ROOM, head, "玩家不在房间内: %d", head.Uid)
	}
	if usr.Uid != head.Uid || usr.ChairId != req.ChairId {
		return uerror.NEW(pb.ErrorCode_PARAM_INVALID, head, "请求参数错误: %v", req)
	}
	if usr.GameInfo.Operate == pb.OperateType_ALL_IN {
		return uerror.NEW(pb.ErrorCode_TEXAS_PLAYER_HAS_ALL_IN, head, "玩家已经allin: %v", usr)
	}
	if usr.GameInfo.Operate == pb.OperateType_FOLD {
		return uerror.NEW(pb.ErrorCode_TEXAS_PLAYER_HAS_FOLD, head, "玩家已经弃牌: %v", usr)
	}
	// 判断筹码是否足够
	if req.OperateType != int32(pb.OperateType_CHECK) && req.OperateType != int32(pb.OperateType_FOLD) {
		switch pb.OperateType(req.OperateType) {
		case pb.OperateType_CALL:
			req.Chip = d.Table.GameData.MaxBetChips - usr.GameInfo.BetChips
		case pb.OperateType_RAISE:
			if req.Chip < d.Table.GameData.MinRaise+d.Table.GameData.MaxBetChips-usr.GameInfo.BetChips {
				req.Chip = d.Table.GameData.MinRaise + d.Table.GameData.MaxBetChips - usr.GameInfo.BetChips
			}
		case pb.OperateType_ALL_IN:
			req.Chip = usr.Chips
		default:
			if usr.Chips < req.Chip || usr.Chips == req.Chip {
				req.OperateType = int32(pb.OperateType_ALL_IN)
				req.Chip = usr.Chips
			} else if req.Chip > d.Table.GameData.MaxBetChips-usr.GameInfo.BetChips {
				req.OperateType = int32(pb.OperateType_RAISE)
			} else if req.Chip == d.Table.GameData.MaxBetChips-usr.GameInfo.BetChips {
				req.OperateType = int32(pb.OperateType_CALL)
			}
		}
	}

	switch pb.OperateType(req.OperateType) {
	case pb.OperateType_CALL:
		if req.Chip != d.Table.GameData.MaxBetChips-usr.GameInfo.BetChips {
			return uerror.NEW(pb.ErrorCode_PARAM_INVALID, head, "请求参数错误: %v", req)
		}
	case pb.OperateType_RAISE:
		if req.Chip < usr.Chips {
			if req.Chip < d.Table.GameData.MinRaise+d.Table.GameData.MaxBetChips-usr.GameInfo.BetChips {
				return uerror.NEW(pb.ErrorCode_TEXAS_CHIPS_NOT_ENOUGH, head, "下注筹码不足: %v", req)
			}
			d.Table.GameData.MinRaise = req.Chip - d.Table.GameData.MaxBetChips + usr.GameInfo.BetChips
		} else {
			minRaise := req.Chip - d.Table.GameData.MaxBetChips + usr.GameInfo.BetChips
			if minRaise > d.Table.GameData.MinRaise {
				d.Table.GameData.MinRaise = minRaise
			}
		}
	}

	// 下注操作
	usr.Chips -= req.Chip
	d.Operate(usr, pb.OperateType(req.OperateType), req.Chip)
	d.Change()

	// 返回客户端
	rsp.Round = d.Table.Round
	rsp.ChairId = req.ChairId
	rsp.OpType = req.OperateType
	rsp.Chip = req.Chip
	rsp.BankRoll = usr.Chips
	rsp.TotalBet = usr.GameInfo.BetChips
	rsp.RoomId = d.RoomId
	return nil
}
