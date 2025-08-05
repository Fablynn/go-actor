package player

import (
	"encoding/json"
	"go-actor/common/pb"
	"go-actor/common/yaml"
	"go-actor/framework/actor"
	"go-actor/library/safe"
	"go-actor/library/uerror"
	"go-actor/server/client/internal/request"
	"go-actor/server/client/internal/stat"
	"reflect"

	"github.com/golang/protobuf/proto"
)

type ClientPlayerMgr struct {
	actor.Actor
	mgr  *actor.ActorMgr
	cfg  *yaml.NodeConfig
	node *pb.Node
	list []uint64
}

func NewClientPlayerMgr(node *pb.Node, cfg *yaml.NodeConfig) *ClientPlayerMgr {
	mgr := new(actor.ActorMgr)
	pl := &ClientPlayer{}
	mgr.Register(pl)
	mgr.ParseFunc(reflect.TypeOf(pl))
	actor.Register(mgr)

	ret := new(ClientPlayerMgr)
	ret.Actor.Register(ret)
	ret.Actor.ParseFunc(reflect.TypeOf(ret))
	ret.Start()
	actor.Register(ret)

	ret.mgr = mgr
	ret.cfg = cfg
	ret.node = node
	return ret
}

func (p *ClientPlayerMgr) Remove(uid uint64) {
	if ac := p.mgr.GetActor(uid); ac != nil {
		p.mgr.DelActor(uid)
		safe.Go(func() {
			ac.Stop()
		})
	}
}

func (p *ClientPlayerMgr) Login(begin, end uint64) {
	for i := begin; i <= end; i++ {
		p.list = append(p.list, i)
	}

	cmdst := stat.NewCmdStat(uint32(pb.CMD_CMDGATE_LOGIN_REQ), p.list...)
	for i := begin; i <= end; i++ {
		pl := NewClientPlayer(p.node, p.cfg, i)
		p.mgr.AddActor(pl)
		pl.SendMsg(&pb.Head{FuncName: "Login"}, cmdst)
	}
	actor.SendMsg(&pb.Head{ActorName: "Statistics", FuncName: "Analysis"}, cmdst)
}

func (p *ClientPlayerMgr) SendCmd(cmd uint32, uid, routeId uint64, msg string) error {
	f, ok := request.Cmds[cmd]
	if !ok {
		return uerror.New(pb.ErrorCode_CMD_NOT_FOUND, "cmd %d not found", cmd)
	}

	req := f()
	if err := json.Unmarshal([]byte(msg), req); err != nil {
		return err
	}
	buf, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	var cmdst *stat.CmdStat
	if uid > 0 {
		cmdst = stat.NewCmdStat(cmd, uid)
		p.mgr.SendMsg(&pb.Head{
			ActorId:   uid,
			Uid:       uid,
			ActorName: "Player",
			FuncName:  "SendCmd",
			SendType:  pb.SendType_POINT,
		}, cmdst, cmd, routeId, buf)
	} else {
		cmdst = stat.NewCmdStat(cmd, p.list...)
		p.mgr.SendMsg(&pb.Head{
			ActorId:   uid,
			ActorName: "Player",
			FuncName:  "SendCmd",
			SendType:  pb.SendType_BROADCAST,
		}, cmdst, cmd, routeId, buf)
	}
	return actor.SendMsg(&pb.Head{ActorName: "Statistics", FuncName: "Analysis"}, cmdst)
}
