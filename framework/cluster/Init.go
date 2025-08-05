package cluster

import (
	"go-actor/common/pb"
	"go-actor/common/yaml"
	"go-actor/framework/internal/bus"
	"go-actor/framework/internal/discovery"
	"go-actor/framework/internal/method"
	"go-actor/framework/internal/node"
	"go-actor/framework/internal/request"
	"go-actor/framework/internal/router"
	"go-actor/library/mlog"

	"github.com/golang/protobuf/proto"
)

var (
	obj *Cluster
)

func Init(nn *pb.Node, srvCfg *yaml.NodeConfig, cfg *yaml.Config) error {
	// 服务注册与发现
	cls := node.New(nn)
	tab := router.New(srvCfg.RouterTTL)
	dis, err := discovery.NewEtcd(cfg.Etcd)
	if err != nil {
		return err
	}
	if err := dis.Watch(cls); err != nil {
		return err
	}
	if err := dis.Register(cls, srvCfg.DiscoveryTTL); err != nil {
		return err
	}

	// 消息中间件
	buss, err := bus.NewNats(cfg.Nats, tab)
	if err != nil {
		return err
	}
	obj = New(cls, tab, buss, dis)
	method.Init(obj.SendResponse)
	return nil
}

func Close() {
	obj.Close()
}

func GetSelf() *pb.Node {
	return obj.GetSelf()
}

func GetAppName() string {
	return obj.GetSelf().Name
}

func SendResponse(head *pb.Head, rsp proto.Message) (err error) {
	err = obj.SendResponse(head, rsp)
	mlog.Trace(head, "SendResponse rsp<%v>|Error<%v>", rsp, err)
	return
}

func SetBroadcastHandler(f func(*pb.Head, []byte)) error {
	return obj.SetBroadcastHandler(f)
}

func SetSendHandler(f func(*pb.Head, []byte)) error {
	return obj.SetSendHandler(f)
}

func SetReplyHandler(f func(*pb.Head, []byte)) error {
	return obj.SetReplyHandler(f)
}

func Broadcast(head *pb.Head, args ...interface{}) error {
	err := obj.Broadcast(head, args...)
	mlog.Trace(head, "Broadcast Args<%v>|Error<%v>", args, err)
	return err
}

func Send(head *pb.Head, args ...interface{}) error {
	err := obj.Send(head, args...)
	mlog.Trace(head, "Send Args<%v>|Error<%v>", args, err)
	return err
}

func SendToClient(head *pb.Head, msg proto.Message, uids ...uint64) error {
	err := obj.SendToClient(head, msg, uids...)
	mlog.Trace(head, "SendToClient Msgs<%v>|Uids<%v>|Error<%v>", msg, uids, err)
	return err
}

func Request(dst interface{}, msg interface{}, rsp proto.Message) error {
	var head *pb.Head
	switch vv := dst.(type) {
	case *pb.NodeRouter:
		head = &pb.Head{Dst: vv}
	case *pb.Head:
		head = vv
	}
	err := obj.Request(head, msg, rsp)
	mlog.Trace(head, "Request Msgs<%v>|Rsp<%v>|Error<%v>", msg, rsp, err)
	return err
}

func Response(head *pb.Head, msg interface{}) error {
	err := obj.Response(head, msg)
	mlog.Trace(head, "Response Rsp<%v>|Error<%v>", msg, err)
	return err
}

// 同步路由
func SendRouter(head *pb.Head, rt pb.RouterType, ids ...uint64) error {
	return obj.SendRouter(head, rt, ids...)
}

// gate服务，默认uid路由, 转发
func SendToGate(head *pb.Head, actorFunc string, args ...interface{}) error {
	head.Dst = request.NewNodeRouter(pb.NodeType_NodeTypeGate, pb.RouterType_UID, head.Uid, 0, actorFunc)
	return Send(head, args...)
}

// game服务，默认uid路由, 转发
func SendToGame(head *pb.Head, actorFunc string, args ...interface{}) error {
	head.Dst = request.NewNodeRouter(pb.NodeType_NodeTypeGame, pb.RouterType_UID, head.Uid, 0, actorFunc)
	return Send(head, args...)
}

// match服务，默认uid路由, 转发
func SendToMatch(head *pb.Head, actorId uint64, actorFunc string, args ...interface{}) error {
	head.Dst = request.NewNodeRouter(pb.NodeType_NodeTypeMatch, pb.RouterType_UID, head.Uid, actorId, actorFunc)
	return Send(head, args...)
}

// room服务，默认roomId路由, 转发
func SendToRoom(head *pb.Head, roomId uint64, actorFunc string, args ...interface{}) error {
	head.Dst = request.NewNodeRouter(pb.NodeType_NodeTypeRoom, pb.RouterType_ROOM_ID, roomId, 0, actorFunc)
	return Send(head, args...)
}

// db服务，默认uid路由, 转发
func SendToDb(head *pb.Head, actorFunc string, args ...interface{}) error {
	head.Dst = request.NewNodeRouter(pb.NodeType_NodeTypeDb, pb.RouterType_UID, head.Uid, 0, actorFunc)
	return Send(head, args...)
}

// db服务，默认roomId路由, 转发
func SendToDbRoomId(head *pb.Head, rid uint64, actorFunc string, args ...interface{}) error {
	head.Dst = request.NewNodeRouter(pb.NodeType_NodeTypeDb, pb.RouterType_ROOM_ID, rid, 0, actorFunc)
	return Send(head, args...)
}

// builder服务，默认随机路由，转发
func SendToBuilder(head *pb.Head, actorFunc string, args ...interface{}) error {
	head.Dst = request.NewNodeRouter(pb.NodeType_NodeTypeBuilder, pb.RouterType_RANDOM_ID, 0, 0, actorFunc)
	return Send(head, args...)
}

// -----------------request-----------------
func RequestToRoom(head *pb.Head, roomId uint64, actorFunc string, msg interface{}, rsp proto.Message) error {
	head.Dst = request.NewNodeRouter(pb.NodeType_NodeTypeRoom, pb.RouterType_ROOM_ID, roomId, 0, actorFunc)
	return Request(head, msg, rsp)
}

func RequestToMatch(head *pb.Head, actorId uint64, actorFunc string, msg interface{}, rsp proto.Message) error {
	head.Dst = request.NewNodeRouter(pb.NodeType_NodeTypeMatch, pb.RouterType_UID, head.Uid, actorId, actorFunc)
	return Request(head, msg, rsp)
}

func RequestToBuilder(actorFunc string, msg interface{}, rsp proto.Message) error {
	return Request(request.NewNodeRouter(pb.NodeType_NodeTypeBuilder, pb.RouterType_RANDOM_ID, 0, 0, actorFunc), msg, rsp)
}
