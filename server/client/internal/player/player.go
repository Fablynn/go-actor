package player

import (
	"fmt"
	"poker_server/common/pb"
	"poker_server/common/yaml"
	"poker_server/framework/actor"
	"poker_server/framework/domain"
	"poker_server/framework/network"
	"poker_server/framework/token"
	"poker_server/library/async"
	"poker_server/library/mlog"
	"poker_server/library/uerror"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

type ClientPlayer struct {
	actor.Actor
	conn domain.INet
	cfg  *yaml.ServerConfig
	cmds map[uint32]func() proto.Message
	node *pb.Node
	uid  uint64
}

func NewClientPlayer(node *pb.Node, cfg *yaml.ServerConfig, uid uint64, cmds map[uint32]func() proto.Message) *ClientPlayer {
	ret := &ClientPlayer{
		node: node,
		cfg:  cfg,
		uid:  uid,
		cmds: cmds,
	}
	ret.Actor.Register(ret)
	ret.SetId(uid)
	ret.Start()
	return ret
}

func (p *ClientPlayer) SendCmd(cmd uint32, routeId uint64, buf []byte) error {
	if routeId <= 0 {
		routeId = p.uid
	}
	head := &pb.Head{
		Src: &pb.NodeRouter{ActorId: routeId},
		Dst: &pb.NodeRouter{
			NodeType: p.node.Type,
			NodeId:   p.node.Id,
			ActorId:  routeId,
		},
		Uid: p.uid,
		Cmd: uint32(cmd),
	}
	return p.conn.Write(&pb.Packet{Head: head, Body: buf})
}

func (p *ClientPlayer) Login() error {
	head := &pb.Head{
		ActorName: "PlayerMgr",
		FuncName:  "Remove",
		Uid:       p.uid,
	}

	// 建立连接
	wsUrl := fmt.Sprintf("ws://%s:%d/ws", p.cfg.Ip, p.cfg.Port)
	ws, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		actor.SendMsg(head, p.uid)
		return err
	}
	p.conn = network.NewSocket(ws, 1024*1024)
	p.conn.Register(&Frame{node: p.node})

	// 设置 session
	tok, err := token.GenToken(&token.Token{Uid: p.uid})
	if err != nil {
		actor.SendMsg(head, p.uid)
		return err
	}

	// 发送登录请求
	buf, _ := proto.Marshal(&pb.GateLoginRequest{Token: tok})
	if err := p.SendCmd(uint32(pb.CMD_GATE_LOGIN_REQUEST), p.uid, buf); err != nil {
		actor.SendMsg(head, p.uid)
		return err
	}

	// 接收登录返回消息
	pack := &pb.Packet{}
	if err := p.conn.Read(pack); err != nil {
		actor.SendMsg(head, p.uid)
		return err
	}
	loginRsp := &pb.GateLoginResponse{}
	if err := proto.Unmarshal(pack.Body, loginRsp); err != nil {
		actor.SendMsg(head, p.uid)
		return err
	}
	if loginRsp.Head != nil {
		return uerror.ToError(loginRsp.Head)
	}

	async.SafeGo(mlog.Errorf, p.loop)
	async.SafeGo(mlog.Errorf, p.keepAlive)
	return nil
}

func (p *ClientPlayer) loop() {
	for {
		pack := &pb.Packet{}
		if err := p.conn.Read(pack); err != nil {
			mlog.Errorf("读取消息失败: %v", err)
			break
		}

		switch pack.Head.Cmd {
		case uint32(pb.CMD_GATE_HEART_RESPONSE):
		default:
			if ff, ok := p.cmds[pack.Head.Cmd]; ok {
				msg := ff()
				if err := proto.Unmarshal(pack.Body, msg); err != nil {
					mlog.Errorf("反序列化失败: %v", err)
					break
				}
				mlog.Infof("[%d] [%s]: %v, rsp:%s", p.uid, pb.CMD(pack.Head.Cmd).String(), pack.Head, msg.String())
			} else {
				mlog.Infof("[%d]: %v, body:%s", p.uid, pack.Head.Cmd, pack.Head, string(pack.Body))
			}
		}
	}
}

func (p *ClientPlayer) keepAlive() {
	// 循环发送心跳
	tt := time.NewTicker(3 * time.Second)
	defer tt.Stop()
	buf, _ := proto.Marshal(&pb.GateHeartRequest{})
	for {
		<-tt.C
		if err := p.SendCmd(uint32(pb.CMD_GATE_HEART_REQUEST), p.uid, buf); err != nil {
			mlog.Errorf("发送心跳包失败: %v", err)
			break
		}
	}
}
