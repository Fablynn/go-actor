/*
* 本代码由cfgtool工具生成，请勿手动修改
 */

package manager

import (
	"go-actor/common/pb"

	"github.com/golang/protobuf/proto"
)

var (
	cmds = make(map[uint32]func() proto.Message)
)

func init() {
	cmds[16777216] = func() proto.Message { return &pb.GateLoginRequest{} }
	cmds[16777217] = func() proto.Message { return &pb.GateLoginResponse{} }
	cmds[33554432] = func() proto.Message { return &pb.GateHeartRequest{} }
	cmds[33554433] = func() proto.Message { return &pb.GateHeartResponse{} }
	cmds[33554434] = func() proto.Message { return &pb.GetBagReq{} }
	cmds[33554435] = func() proto.Message { return &pb.GetBagRsp{} }
	cmds[50331648] = func() proto.Message { return &pb.TexasRoomListReq{} }
	cmds[50331649] = func() proto.Message { return &pb.TexasRoomListRsp{} }
	cmds[50331650] = func() proto.Message { return &pb.RummyRoomListReq{} }
	cmds[50331651] = func() proto.Message { return &pb.RummyRoomListRsp{} }
	cmds[50331652] = func() proto.Message { return &pb.SngJoinRoomReq{} }
	cmds[50331653] = func() proto.Message { return &pb.SngJoinRoomRsp{} }
	cmds[67108864] = func() proto.Message { return &pb.TexasEventNotify{} }
	cmds[67108866] = func() proto.Message { return &pb.TexasJoinRoomReq{} }
	cmds[67108867] = func() proto.Message { return &pb.TexasJoinRoomRsp{} }
	cmds[67108868] = func() proto.Message { return &pb.TexasQuitRoomReq{} }
	cmds[67108869] = func() proto.Message { return &pb.TexasQuitRoomRsp{} }
	cmds[67108870] = func() proto.Message { return &pb.TexasSitDownReq{} }
	cmds[67108871] = func() proto.Message { return &pb.TexasSitDownRsp{} }
	cmds[67108872] = func() proto.Message { return &pb.TexasStandUpReq{} }
	cmds[67108873] = func() proto.Message { return &pb.TexasStandUpRsp{} }
	cmds[67108880] = func() proto.Message { return &pb.TexasBuyInReq{} }
	cmds[67108881] = func() proto.Message { return &pb.TexasBuyInRsp{} }
	cmds[67108882] = func() proto.Message { return &pb.TexasDoBetReq{} }
	cmds[67108883] = func() proto.Message { return &pb.TexasDoBetRsp{} }
	cmds[67108884] = func() proto.Message { return &pb.RummyJoinRoomReq{} }
	cmds[67108885] = func() proto.Message { return &pb.RummyJoinRoomRsp{} }
	cmds[67108886] = func() proto.Message { return &pb.RummyEventNotify{} }
	cmds[67108888] = func() proto.Message { return &pb.RummyQuitRoomReq{} }
	cmds[67108889] = func() proto.Message { return &pb.RummyQuitRoomRsp{} }
	cmds[67108890] = func() proto.Message { return &pb.TexasChangeRoomReq{} }
	cmds[67108891] = func() proto.Message { return &pb.TexasChangeRoomRsp{} }
	cmds[67108896] = func() proto.Message { return &pb.RummySaveCardGroupReq{} }
	cmds[67108897] = func() proto.Message { return &pb.RummySaveCardGroupRsp{} }
	cmds[67108898] = func() proto.Message { return &pb.RummyOprCardReq{} }
	cmds[67108899] = func() proto.Message { return &pb.RummyOprCardRsp{} }
	cmds[67108900] = func() proto.Message { return &pb.RummyFixCardReq{} }
	cmds[67108901] = func() proto.Message { return &pb.RummyFixCardRsp{} }
	cmds[67108902] = func() proto.Message { return &pb.RummyReadyRoomReq{} }
	cmds[67108903] = func() proto.Message { return &pb.RummyReadyRoomRsp{} }
	cmds[67108904] = func() proto.Message { return &pb.RummyGetOutCardsReq{} }
	cmds[67108905] = func() proto.Message { return &pb.RummyGetOutCardsRsp{} }
	cmds[67108906] = func() proto.Message { return &pb.TexasStatisticsReq{} }
	cmds[67108907] = func() proto.Message { return &pb.TexasStatisticsRsp{} }
	cmds[67108908] = func() proto.Message { return &pb.SngRankReq{} }
	cmds[67108909] = func() proto.Message { return &pb.SngRankRsp{} }
	cmds[67108912] = func() proto.Message { return &pb.RummyChangeRoomReq{} }
	cmds[67108913] = func() proto.Message { return &pb.RummyChangeRoomRsp{} }
	cmds[83886080] = func() proto.Message { return &pb.TexasGameReportReq{} }
	cmds[83886081] = func() proto.Message { return &pb.TexasGameReportRsp{} }
	cmds[83886082] = func() proto.Message { return &pb.RummyMatchSelectReq{} }
	cmds[83886083] = func() proto.Message { return &pb.RummyMatchSelectRsp{} }
	cmds[83886084] = func() proto.Message { return &pb.GetUserInfoReq{} }
	cmds[83886085] = func() proto.Message { return &pb.GetUserInfoRsp{} }
	cmds[83886086] = func() proto.Message { return &pb.GetTexasGameReportReq{} }
	cmds[83886087] = func() proto.Message { return &pb.GetTexasGameReportRsp{} }
}
