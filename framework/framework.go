package framework

import (
	"poker_server/common/pb"
	"poker_server/common/yaml"
	"poker_server/framework/actor"
	"poker_server/framework/internal/service"
	"poker_server/library/mlog"
	"poker_server/library/uerror"
	"poker_server/library/util"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
)

var (
	core *service.Service
)

func Init(node *pb.Node, server *yaml.ServerConfig, cfg *yaml.Config) (err error) {
	core, err = service.NewService(node, server, cfg)
	if err != nil {
		return
	}
	actor.Init(node, SendResponse)
	return
}

func InitDefault(node *pb.Node, server *yaml.ServerConfig, cfg *yaml.Config) (err error) {
	core, err = service.NewService(node, server, cfg)
	if err != nil {
		return
	}
	actor.Init(node, SendResponse)
	core.RegisterBroadcastHandler(DefaultBroadcastHandler)
	core.RegisterSendHandler(DefaultSendHandler)
	core.RegisterReplyHandler(DefaultReplyHandler)
	return
}

func Close() error {
	return core.Close()
}

func GetSelf() *pb.Node {
	return core.GetNode()
}

func GetEnvType() pb.EnvType {
	return core.GetEnvType()
}

// 跨服务发消息
func Send(head *pb.Head, args ...interface{}) error {
	return core.Send(head, args...)
}

// 跨服务类型广播
func Broadcast(head *pb.Head, args ...interface{}) error {
	return core.Broadcast(head, args...)
}

// 同步请求
func Request(head *pb.Head, msg interface{}, reply proto.Message) error {
	return core.Request(head, msg, reply)
}

func Response(head *pb.Head, msg interface{}) error {
	return core.Response(head, msg)
}

// 发送到客户端
func SendToClient(head *pb.Head, msg proto.Message) error {
	return core.SendToClient(head, msg)
}

// 通知客户端
func NotifyToClient(uids []uint64, head *pb.Head, msg proto.Message) error {
	return core.NotifyToClient(uids, head, msg)
}

// 注册消息处理函数
func RegisterBroadcastHandler(f func(*pb.Head, []byte)) error {
	return core.RegisterBroadcastHandler(f)
}

// 注册消息处理函数
func RegisterSendHandler(f func(*pb.Head, []byte)) error {
	return core.RegisterSendHandler(f)
}

// 注册消息处理函数
func RegisterReplyHandler(f func(*pb.Head, []byte)) error {
	return core.RegisterReplyHandler(f)
}

// 默认内网消息处理器
func DefaultSendHandler(head *pb.Head, buf []byte) {
	head.ActorName = head.Dst.ActorName
	head.FuncName = head.Dst.FuncName
	head.ActorId = head.Dst.ActorId
	if err := actor.Send(head, buf); err != nil {
		mlog.Errorf("跨服务调用actor错误: %v head: %v", err, head)
	}
}

// 默认内网消息处理器
func DefaultReplyHandler(head *pb.Head, buf []byte) {
	head.ActorName = head.Dst.ActorName
	head.FuncName = head.Dst.FuncName
	head.ActorId = head.Dst.ActorId
	if err := actor.Send(head, buf); err != nil {
		mlog.Errorf("跨服务调用错误: %v", err)
	}
}

// 默认内网广播消息处理器
func DefaultBroadcastHandler(head *pb.Head, buf []byte) {
	head.ActorName = head.Dst.ActorName
	head.FuncName = head.Dst.FuncName
	head.ActorId = head.Dst.ActorId
	if err := actor.Send(head, buf); err != nil {
		mlog.Errorf("跨服务调用错误: %v", err)
	}
}

func SendResponse(head *pb.Head, rsp proto.Message) error {
	// 同步请求
	if len(head.Reply) > 0 {
		return Response(head, rsp)
	}

	// cmd请求
	head.Src, head.Dst = head.Dst, head.Src
	if head.Cmd > 0 {
		return core.SendToClient(head, rsp)
	}

	// 跨服务异步请求
	if len(head.Src.ActorName) <= 0 || len(head.Src.FuncName) <= 0 {
		return uerror.New(1, pb.ErrorCode_PARAM_INVALID, "head is nil, head:%v", head)
	}
	return core.Send(head, rsp)
}

func StopAutoSendToClient(head *pb.Head) {
	atomic.AddInt32(&head.Reference, 1)
}

func NewSrcRouter(rt pb.RouterType, id uint64, fs ...string) *pb.NodeRouter {
	node := core.GetNode()
	return &pb.NodeRouter{
		NodeType:   node.Type,
		NodeId:     node.Id,
		RouterType: rt,
		ActorId:    id,
		ActorName:  util.Index[string](fs, 0, ""),
		FuncName:   util.Index[string](fs, 1, ""),
	}
}

func NewGateRouter(id uint64, fs ...string) *pb.NodeRouter {
	return &pb.NodeRouter{
		NodeType:   pb.NodeType_NodeTypeGate,
		RouterType: pb.RouterType_RouterTypeUid,
		ActorId:    id,
		ActorName:  util.Index[string](fs, 0, ""),
		FuncName:   util.Index[string](fs, 1, ""),
	}
}

func NewGameRouter(id uint64, fs ...string) *pb.NodeRouter {
	return &pb.NodeRouter{
		NodeType:   pb.NodeType_NodeTypeGame,
		RouterType: pb.RouterType_RouterTypeUid,
		ActorId:    id,
		ActorName:  util.Index[string](fs, 0, ""),
		FuncName:   util.Index[string](fs, 1, ""),
	}
}

func NewRoomRouter(id uint64, fs ...string) *pb.NodeRouter {
	return &pb.NodeRouter{
		NodeType:   pb.NodeType_NodeTypeRoom,
		RouterType: pb.RouterType_RouterTypeRoomId,
		ActorId:    id,
		ActorName:  util.Index[string](fs, 0, ""),
		FuncName:   util.Index[string](fs, 1, ""),
	}
}

func NewBuilderRouter(id uint64, fs ...string) *pb.NodeRouter {
	return &pb.NodeRouter{
		NodeType:   pb.NodeType_NodeTypeBuilder,
		RouterType: pb.RouterType_RouterTypeGeneratorType,
		ActorId:    id,
		ActorName:  util.Index[string](fs, 0, ""),
		FuncName:   util.Index[string](fs, 1, ""),
	}
}

func NewDbRouter(id uint64, fs ...string) *pb.NodeRouter {
	return &pb.NodeRouter{
		NodeType:   pb.NodeType_NodeTypeDb,
		RouterType: pb.RouterType_RouterTypeDataType,
		ActorId:    id,
		ActorName:  util.Index[string](fs, 0, ""),
		FuncName:   util.Index[string](fs, 1, ""),
	}
}

func NewMatchRouter(id uint64, fs ...string) *pb.NodeRouter {
	return &pb.NodeRouter{
		NodeType:   pb.NodeType_NodeTypeMatch,
		RouterType: pb.RouterType_RouterTypeGameType,
		ActorId:    id,
		ActorName:  util.Index[string](fs, 0, ""),
		FuncName:   util.Index[string](fs, 1, ""),
	}
}

func NewHead(dst *pb.NodeRouter, rt pb.RouterType, id uint64, fs ...string) *pb.Head {
	return &pb.Head{Dst: dst, Src: NewSrcRouter(rt, id, fs...)}
}

func SwapToDb(head *pb.Head, id uint64, fs ...string) *pb.Head {
	head.Dst, head.Src = head.Src, head.Dst
	head.Dst.NodeType = pb.NodeType_NodeTypeDb
	head.Dst.RouterType = pb.RouterType_RouterTypeDataType
	head.Dst.ActorId = id
	head.Dst.ActorName = util.Index[string](fs, 0, "")
	head.Dst.FuncName = util.Index[string](fs, 1, "")
	return head
}

func SwapToGame(head *pb.Head, id uint64, fs ...string) *pb.Head {
	head.Dst, head.Src = head.Src, head.Dst
	head.Dst.NodeType = pb.NodeType_NodeTypeGame
	head.Dst.RouterType = pb.RouterType_RouterTypeUid
	head.Dst.ActorId = id
	head.Dst.ActorName = util.Index[string](fs, 0, "")
	head.Dst.FuncName = util.Index[string](fs, 1, "")
	return head
}

func SwapToRoom(head *pb.Head, id uint64, fs ...string) *pb.Head {
	head.Dst, head.Src = head.Src, head.Dst
	head.Dst.NodeType = pb.NodeType_NodeTypeRoom
	head.Dst.RouterType = pb.RouterType_RouterTypeRoomId
	head.Dst.ActorId = id
	head.Dst.ActorName = util.Index[string](fs, 0, "")
	head.Dst.FuncName = util.Index[string](fs, 1, "")
	return head
}

func SwapToGate(head *pb.Head, id uint64, fs ...string) *pb.Head {
	head.Dst, head.Src = head.Src, head.Dst
	head.Dst.NodeType = pb.NodeType_NodeTypeGate
	head.Dst.RouterType = pb.RouterType_RouterTypeUid
	head.Dst.ActorId = id
	head.Dst.ActorName = util.Index[string](fs, 0, "")
	head.Dst.FuncName = util.Index[string](fs, 1, "")
	return head
}

func SwapToMatch(head *pb.Head, id uint64, fs ...string) *pb.Head {
	head.Dst, head.Src = head.Src, head.Dst
	head.Dst.NodeType = pb.NodeType_NodeTypeMatch
	head.Dst.RouterType = pb.RouterType_RouterTypeDataType
	head.Dst.ActorId = id
	head.Dst.ActorName = util.Index[string](fs, 0, "")
	head.Dst.FuncName = util.Index[string](fs, 1, "")
	return head
}
