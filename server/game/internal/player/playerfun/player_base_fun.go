package playerfun

import (
	"go-actor/common/pb"
	"go-actor/common/room_util"
	"go-actor/framework"
	"go-actor/library/uerror"
	"go-actor/server/game/internal/player/domain"

	"github.com/golang/protobuf/proto"
)

type PlayerBaseFun struct {
	*PlayerFun
	data *pb.PlayerDataBase
}

func NewPlayerBaseFun(fun *PlayerFun) domain.IPlayerFun {
	return &PlayerBaseFun{PlayerFun: fun}
}

func (d *PlayerBaseFun) Load(msg *pb.PlayerData) error {
	if msg == nil {
		return uerror.New(1, pb.ErrorCode_PARAM_INVALID, "玩家基础数据为空")
	}
	if msg == nil || msg.Base == nil {
		return uerror.New(1, pb.ErrorCode_PARAM_INVALID, "玩家基础数据为空")
	}
	d.data = msg.Base
	return nil
}

func (d *PlayerBaseFun) Save(data *pb.PlayerData) error {
	if data == nil {
		return uerror.New(1, pb.ErrorCode_PARAM_INVALID, "玩家数据为空")
	}
	buf, _ := proto.Marshal(d.data)
	newBase := &pb.PlayerDataBase{}
	proto.Unmarshal(buf, newBase)
	data.Base = newBase
	return nil
}

// QueryPlayerData for reconnect
func (d *PlayerBaseFun) QueryPlayerData(head *pb.Head, req *pb.QueryPlayerDataReq, rsp *pb.QueryPlayerDataRsp) error {
	rsp.Data = d.data
	if d.data.RoomInfo != nil {
		rsp.MatchType, rsp.GameType, rsp.CoinType = room_util.TexasRoomIdTo(d.data.RoomInfo.RoomId)
	}
	return nil
}

func (d *PlayerBaseFun) LoadComplate() error {
	return nil
}

func (d *PlayerBaseFun) UpdatePlayerInfo(info *pb.PlayerInfo) {
	if info == nil {
		return
	}
	d.data.PlayerInfo = info
}

func (d *PlayerBaseFun) GetPlayerInfo() *pb.PlayerInfo {
	return d.data.PlayerInfo
}

func (d *PlayerBaseFun) GetRoomInfo() *pb.PlayerRoomInfo {
	return d.data.RoomInfo
}

func (d *PlayerBaseFun) GetTexasRealRoomId(head *pb.Head, roomId uint64) (uint64, error) {
	matchType, gameType, coinType := room_util.TexasRoomIdTo(roomId)
	if d.data.RoomInfo == nil || gameType == d.data.RoomInfo.GameType && d.data.RoomInfo.RoomId == roomId {
		return roomId, nil
	}
	if d.data.RoomInfo.GameType != gameType {
		return 0, uerror.NEW(pb.ErrorCode_GAME_PLAYER_IN_OTHER_GAME, head, "玩家已在其他游戏中,无法加入德州扑克房间")
	}

	// 断线重连
	newrsp := &pb.HasRoomRsp{}
	newreq := &pb.HasRoomReq{RoomId: d.data.RoomInfo.RoomId}
	var dst *pb.NodeRouter
	switch matchType {
	case pb.MatchType_MatchTypeNone:
		dst = framework.NewMatchRouter(uint64(matchType)<<32|uint64(gameType)<<16|uint64(coinType), "MatchTexasRoom", "HasRoomReq")
	}
	if err := framework.Request(&pb.Head{Dst: dst}, newreq, newrsp); err != nil {
		return 0, err
	} else if newrsp.Head != nil {
		return 0, uerror.ToError(newrsp.Head)
	}
	if newrsp.IsExist {
		return d.data.RoomInfo.RoomId, nil
	}
	return roomId, nil
}

func (d *PlayerBaseFun) TexasJoinRoom(roomId uint64) {
	_, gameType, _ := room_util.TexasRoomIdTo(roomId)
	d.data.RoomInfo = &pb.PlayerRoomInfo{GameType: gameType, RoomId: roomId}
}

func (d *PlayerBaseFun) TexasQuitRoom() {
	d.data.RoomInfo = nil
	d.Change()
}
