package test

import (
	"go-actor/common/pb"
	"go-actor/framework/mock"
	"testing"
)

func TestTexas(t *testing.T) {
	uid1, uid2, uid3 := uint64(144), uint64(145), uint64(146)
	roomId := uint64(4311744517)
	// -----------加入房间------------
	t.Run("JoinRoom1", func(t *testing.T) {
		mock.Request(uid1, roomId, pb.CMD_TEXAS_JOIN_ROOM_REQ, &pb.TexasJoinRoomReq{RoomId: roomId})
	})
	t.Run("JoinRoom2", func(t *testing.T) {
		mock.Request(uid2, roomId, pb.CMD_TEXAS_JOIN_ROOM_REQ, &pb.TexasJoinRoomReq{RoomId: roomId})
	})
	t.Run("JoinRoom3", func(t *testing.T) {
		mock.Request(uid3, roomId, pb.CMD_TEXAS_JOIN_ROOM_REQ, &pb.TexasJoinRoomReq{RoomId: roomId})
	})

	// -----------换房间------------
	t.Run("ChangeRoom1", func(t *testing.T) {
		mock.Request(uid1, roomId, pb.CMD_TEXAS_CHANGE_ROOM_REQ, &pb.TexasChangeRoomReq{RoomId: roomId})
	})
	t.Run("ChangeRoom2", func(t *testing.T) {
		mock.Request(uid2, roomId, pb.CMD_TEXAS_CHANGE_ROOM_REQ, &pb.TexasChangeRoomReq{RoomId: roomId})
	})

	// -----------退出房间------------
	t.Run("QuitRoom1", func(t *testing.T) {
		mock.Request(uid1, roomId, pb.CMD_TEXAS_QUIT_ROOM_REQ, &pb.TexasQuitRoomReq{RoomId: roomId})
	})
	t.Run("QuitRoom2", func(t *testing.T) {
		mock.Request(uid2, roomId, pb.CMD_TEXAS_QUIT_ROOM_REQ, &pb.TexasQuitRoomReq{RoomId: roomId})
	})

	// -----------买入------------
	t.Run("BuyIn1", func(t *testing.T) {
		mock.Request(uid1, roomId, pb.CMD_TEXAS_BUY_IN_REQ, &pb.TexasBuyInReq{RoomId: roomId, Chip: 100000, CoinType: 1})
	})
	t.Run("BuyIn2", func(t *testing.T) {
		mock.Request(uid2, roomId, pb.CMD_TEXAS_BUY_IN_REQ, &pb.TexasBuyInReq{RoomId: roomId, Chip: 100000, CoinType: 1})
	})
	t.Run("BuyIn3", func(t *testing.T) {
		mock.Request(uid3, roomId, pb.CMD_TEXAS_BUY_IN_REQ, &pb.TexasBuyInReq{RoomId: roomId, Chip: 100000, CoinType: 1})
	})

	// -----------坐下------------
	t.Run("SitDown1", func(t *testing.T) {
		mock.Request(uid1, roomId, pb.CMD_TEXAS_SIT_DOWN_REQ, &pb.TexasSitDownReq{RoomId: roomId, ChairId: 1})
	})
	t.Run("SitDown2", func(t *testing.T) {
		mock.Request(uid2, roomId, pb.CMD_TEXAS_SIT_DOWN_REQ, &pb.TexasSitDownReq{RoomId: roomId, ChairId: 2})
	})
	t.Run("SitDown3", func(t *testing.T) {
		mock.Request(uid3, roomId, pb.CMD_TEXAS_SIT_DOWN_REQ, &pb.TexasSitDownReq{RoomId: roomId, ChairId: 3})
	})

	// ------------站起------------
	t.Run("StandUp1", func(t *testing.T) {
		mock.Request(uid1, roomId, pb.CMD_TEXAS_STAND_UP_REQ, &pb.TexasStandUpReq{RoomId: roomId, ChairId: 1})
	})
	t.Run("StandUp2", func(t *testing.T) {
		mock.Request(uid2, roomId, pb.CMD_TEXAS_STAND_UP_REQ, &pb.TexasStandUpReq{RoomId: roomId, ChairId: 2})
	})
	t.Run("StandUp3", func(t *testing.T) {
		mock.Request(uid3, roomId, pb.CMD_TEXAS_STAND_UP_REQ, &pb.TexasStandUpReq{RoomId: roomId, ChairId: 3})
	})

	// ------------下注请求------------
	t.Run("Bet1", func(t *testing.T) {
		mock.Request(uid1, roomId, pb.CMD_TEXAS_DO_BET_REQ, &pb.TexasDoBetReq{Chip: 800, ChairId: 1, RoomId: roomId, OperateType: int32(pb.OperateType_BET)})
	})
	t.Run("Bet2", func(t *testing.T) {
		mock.Request(uid2, roomId, pb.CMD_TEXAS_DO_BET_REQ, &pb.TexasDoBetReq{Chip: 800, ChairId: 2, RoomId: roomId, OperateType: int32(pb.OperateType_BET)})
	})
	t.Run("Bet3", func(t *testing.T) {
		mock.Request(uid3, roomId, pb.CMD_TEXAS_DO_BET_REQ, &pb.TexasDoBetReq{Chip: 800, ChairId: 3, RoomId: roomId, OperateType: int32(pb.OperateType_BET)})
	})
}
