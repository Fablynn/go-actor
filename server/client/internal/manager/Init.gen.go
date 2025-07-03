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
	cmds[50331648] = func() proto.Message { return &pb.TexasRoomListReq{} }
	cmds[50331649] = func() proto.Message { return &pb.TexasRoomListRsp{} }
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
	cmds[67108890] = func() proto.Message { return &pb.TexasChangeRoomReq{} }
	cmds[67108891] = func() proto.Message { return &pb.TexasChangeRoomRsp{} }
}
