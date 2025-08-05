package request

import (
	"fmt"
	"go-actor/common/pb"
	"strings"
)

var (
	cmds = make(map[pb.CMD]*CmdInfo)
)

type CmdInfo struct {
	cmd pb.CMD
	nt  pb.NodeType
	rt  pb.RouterType
	af  string
}

func RegisterCmd(nt pb.NodeType, rt pb.RouterType, cmd pb.CMD, actorFunc string) {
	pos := strings.Index(actorFunc, ".")
	if pos <= 0 {
		panic(fmt.Sprintf("Actor接口注册错误%s", actorFunc))
	}
	cmds[cmd] = &CmdInfo{cmd, nt, rt, actorFunc}
}

func NewCmdRouter(cmd uint32, routerId, actorId uint64) *pb.NodeRouter {
	if info, ok := cmds[pb.CMD(cmd)]; ok {
		return &pb.NodeRouter{
			NodeType:   info.nt,
			RouterType: info.rt,
			RouterId:   routerId,
			ActorFunc:  info.af,
			ActorId:    actorId,
		}
	}
	return nil
}

func ParseActorFunc(actorFunc string) (aname, fname string) {
	if pos := strings.Index(actorFunc, "."); pos > 0 {
		aname = actorFunc[:pos]
		fname = actorFunc[pos+1:]
	} else {
		aname = actorFunc
	}
	return
}

func get(nn interface{}) (nt pb.NodeType, id int32, ok bool) {
	switch vv := nn.(type) {
	case *pb.Node:
		nt = vv.Type
		id = vv.Id
		ok = true
	case pb.NodeType:
		nt = vv
		ok = true
	}
	return
}

func NewNodeRouter(nn interface{}, routerType pb.RouterType, routerId, actorId uint64, actorFunc string) *pb.NodeRouter {
	nt, id, ok := get(nn)
	if !ok {
		return nil
	}
	return &pb.NodeRouter{
		NodeType:   nt,
		NodeId:     id,
		RouterType: routerType,
		RouterId:   routerId,
		ActorId:    actorId,
		ActorFunc:  actorFunc,
	}
}

func CopyTo(head *pb.Head, nt pb.NodeType, routerType pb.RouterType, routerId, actorId uint64, actorFunc string) *pb.Head {
	return &pb.Head{
		SendType: head.SendType,
		Src:      head.Src,
		Dst:      NewNodeRouter(nt, routerType, routerId, actorId, actorFunc),
		Uid:      head.Uid,
		Seq:      head.Seq,
		Cmd:      head.Cmd,
		Reply:    head.Reply,
	}
}
