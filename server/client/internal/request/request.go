package request

import (
	"go-actor/common/pb"

	"github.com/golang/protobuf/proto"
)

var (
	Cmds = make(map[uint32]func() proto.Message)
)

func init() {
	Cmds[1] = func() proto.Message { return &pb.GateLoginRequest{} }
	Cmds[2] = func() proto.Message { return &pb.GateLoginResponse{} }
	Cmds[10001] = func() proto.Message { return &pb.KickPlayerNotify{} }
	Cmds[3] = func() proto.Message { return &pb.GateHeartRequest{} }
	Cmds[4] = func() proto.Message { return &pb.GateHeartResponse{} }
	Cmds[5] = func() proto.Message { return &pb.GetBagReq{} }
	Cmds[6] = func() proto.Message { return &pb.GetBagRsp{} }
}
