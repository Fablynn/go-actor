package manager

import (
	"encoding/json"
	"poker_server/common/pb"
	"poker_server/common/yaml"
	"poker_server/framework/actor"
	"poker_server/library/async"
	"poker_server/library/mlog"
	"poker_server/server/client/internal/player"
	"reflect"

	"github.com/golang/protobuf/proto"
)

type ClientPlayerMgr struct {
	actor.Actor
	mgr  *actor.ActorMgr
	cfg  *yaml.ServerConfig
	node *pb.Node
}

func NewClientPlayerMgr(node *pb.Node, cfg *yaml.ServerConfig) *ClientPlayerMgr {
	mgr := new(actor.ActorMgr)
	pl := &player.ClientPlayer{}
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
		async.SafeGo(mlog.Errorf, func() {
			ac.Stop()
		})
	}
}

func (p *ClientPlayerMgr) Login(begin, end uint64) {
	for i := begin; i <= end; i++ {
		pl := player.NewClientPlayer(p.node, p.cfg, i, cmds)
		p.mgr.AddActor(pl)
		pl.SendMsg(&pb.Head{FuncName: "Login"})
	}
}

func (p *ClientPlayerMgr) SendCmd(cmd uint32, uid, routeId uint64, msg string) {
	f, ok := cmds[cmd]
	if !ok {
		mlog.Errorf("cmd %d not found", cmd)
		return
	}
	req := f()
	if err := json.Unmarshal([]byte(msg), req); err != nil {
		mlog.Errorf("unmarshal error: %v", err)
		return
	}
	buf, err := proto.Marshal(req)
	if err != nil {
		mlog.Errorf("marshal error: %v", err)
		return
	}
	sendType := pb.SendType_BROADCAST
	if uid > 0 {
		sendType = pb.SendType_POINT
	}
	p.mgr.SendMsg(&pb.Head{ActorId: uid, Uid: uid, ActorName: "Player", FuncName: "SendCmd", SendType: sendType}, cmd, routeId, buf)
}
