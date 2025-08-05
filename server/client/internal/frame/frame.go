package frame

import (
	"encoding/binary"
	"go-actor/common/pb"
	"go-actor/framework"
)

type WsPacket struct {
	Cmd     uint32 // 消息id
	Uid     uint64 // 玩家uid
	RouteId uint64 // 路由 id
	Seq     uint32 // 序列号
	Version uint32 // 版本号
	Extra   uint32 // 扩展字段
	Body    []byte // 消息体
}

type Frame struct {
	node *pb.Node
}

func New(nn *pb.Node) *Frame {
	return &Frame{node: nn}
}

// 最大包长限制
func (d *Frame) GetSize(pac *pb.Packet) int {
	return 32 + len(pac.Body)
}

// 解码数据包
func (d *Frame) Decode(buf []byte, msg *pb.Packet) error {
	pack := &WsPacket{}
	pos := 0
	pack.Cmd = binary.BigEndian.Uint32(buf[pos:])
	pos += 4
	pack.Uid = binary.BigEndian.Uint64(buf[pos:])
	pos += 8
	pack.RouteId = binary.BigEndian.Uint64(buf[pos:])
	pos += 8
	pack.Seq = binary.BigEndian.Uint32(buf[pos:])
	pos += 4
	pack.Version = binary.BigEndian.Uint32(buf[pos:])
	pos += 4
	pack.Extra = binary.BigEndian.Uint32(buf[pos:])
	pos += 4
	pack.Body = buf[pos:]
	// 确当路由节点
	msg.Body = pack.Body
	msg.Head = &pb.Head{
		Dst: framework.NewCmdRouter(pack.Cmd-pack.Cmd%2, pack.RouteId, 0),
		Uid: pack.Uid,
		Cmd: pack.Cmd,
		Seq: pack.Seq,
	}
	return nil
}

func (d *Frame) Encode(pack *pb.Packet, buf []byte) error {
	// 组包
	pos := 0
	binary.BigEndian.PutUint32(buf[pos:], pack.Head.Cmd) // cmd
	pos += 4
	binary.BigEndian.PutUint64(buf[pos:], pack.Head.Uid) // uid
	pos += 8
	binary.BigEndian.PutUint64(buf[pos:], pack.Head.Src.ActorId) // router id
	pos += 8
	binary.BigEndian.PutUint32(buf[pos:], pack.Head.Seq) // seq
	pos += 4
	binary.BigEndian.PutUint32(buf[pos:], 0) // version
	pos += 4
	binary.BigEndian.PutUint32(buf[pos:], 0) // extra
	pos += 4
	copy(buf[pos:], pack.Body)
	return nil
}
