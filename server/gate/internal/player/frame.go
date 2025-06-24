package player

import (
	"encoding/binary"
	"poker_server/common/config/repository/router_config"
	"poker_server/common/pb"
	"poker_server/framework"
	"poker_server/library/uerror"
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

type Frame struct{}

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
	cfg := router_config.MGetCmd(pack.Cmd - pack.Cmd%2)
	if cfg == nil {
		return uerror.New(1, pb.ErrorCode_CMD_NOT_FOUND, "cmd未注册, 路由配置不存在: %d", pack.Cmd)
	}
	if msg.Head == nil {
		msg.Head = &pb.Head{}
	}
	msg.Head.Src = framework.NewSrcRouter(pb.RouterType_RouterTypeUid, pack.Uid, "Player", "SendToClient")
	msg.Head.Dst = &pb.NodeRouter{NodeType: cfg.NodeType, ActorName: cfg.ActorName, FuncName: cfg.FuncName, RouterType: cfg.RouterType}
	if msg.Head.Dst.NodeType == msg.Head.Src.NodeType {
		msg.Head.Dst.ActorId = pack.Uid
	} else {
		msg.Head.Dst.ActorId = pack.RouteId
	}
	msg.Head.SendType = pb.SendType_POINT
	msg.Head.Uid = pack.Uid
	msg.Head.Cmd = pack.Cmd
	msg.Head.Seq = pack.Seq
	msg.Body = pack.Body
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
