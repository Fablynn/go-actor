package framework

import (
	"go-actor/common/pb"
	"go-actor/framework/cluster"
	"go-actor/framework/internal/request"
	"go-actor/library/util"
	"sync/atomic"

	"github.com/spf13/cast"
)

func StopAutoSendToClient(head *pb.Head) {
	atomic.AddUint32(&head.Reference, 1)
}

func RegisterCmd(nt pb.NodeType, rt pb.RouterType, cmd pb.CMD, actorFunc string) {
	request.RegisterCmd(nt, rt, cmd, actorFunc)
}

func NewCmdRouter(cmd uint32, routerId, actorId uint64) *pb.NodeRouter {
	return request.NewCmdRouter(cmd, routerId, actorId)
}

func NewSrcRouter(routerType pb.RouterType, routerId uint64, args ...interface{}) *pb.NodeRouter {
	actorId := cast.ToUint64(util.Index[interface{}](args, 0, 0))
	actorFunc := cast.ToString(util.Index[interface{}](args, 1, ""))
	return request.NewNodeRouter(cluster.GetSelf(), routerType, routerId, actorId, actorFunc)
}

func ParseActorFunc(actorFunc string) (aname, fname string) {
	return request.ParseActorFunc(actorFunc)
}

func NewNodeRouter(nn interface{}, routerType pb.RouterType, routerId, actorId uint64, actorFunc string) *pb.NodeRouter {
	return request.NewNodeRouter(nn, routerType, routerId, actorId, actorFunc)
}

// uid
func NewGameRouter(uid uint64, actorFunc string) *pb.NodeRouter {
	return request.NewNodeRouter(pb.NodeType_NodeTypeGame, pb.RouterType_UID, uid, 0, actorFunc)
}

func NewGateRouter(uid uint64, actorFunc string) *pb.NodeRouter {
	return request.NewNodeRouter(pb.NodeType_NodeTypeGate, pb.RouterType_UID, uid, 0, actorFunc)
}

func NewRoomRouter(rid uint64, actorFunc string) *pb.NodeRouter {
	return request.NewNodeRouter(pb.NodeType_NodeTypeRoom, pb.RouterType_ROOM_ID, rid, 0, actorFunc)
}
