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
}
