package player

import (
	"bytes"
	"fmt"
	"go-actor/common/pb"
	"go-actor/common/token"
	"go-actor/common/yaml"
	"go-actor/framework/actor"
	"go-actor/framework/define"
	"go-actor/framework/network"
	"go-actor/library/mlog"
	"go-actor/library/safe"
	"go-actor/library/uerror"
	"go-actor/server/client/internal/frame"
	"go-actor/server/client/internal/request"
	"go-actor/server/client/internal/stat"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

var (
	pool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	}
)

func get() *bytes.Buffer {
	return pool.Get().(*bytes.Buffer)
}

func put(val *bytes.Buffer) {
	pool.Put(val)
}

type ClientPlayer struct {
	actor.Actor
	cmds map[uint32]*atomic.Pointer[stat.CmdStat]
	conn define.INet
	cfg  *yaml.NodeConfig
	node *pb.Node
	uid  uint64
}

func NewClientPlayer(node *pb.Node, cfg *yaml.NodeConfig, uid uint64) *ClientPlayer {
	ret := &ClientPlayer{
		cmds: make(map[uint32]*atomic.Pointer[stat.CmdStat]),
		node: node,
		cfg:  cfg,
		uid:  uid,
	}
	for cmd := range request.Cmds {
		ret.cmds[cmd] = new(atomic.Pointer[stat.CmdStat])
	}
	ret.Actor.Register(ret)
	ret.SetId(uid)
	ret.Start()
	return ret
}

func (p *ClientPlayer) Login(st *stat.CmdStat) error {
	head := &pb.Head{ActorName: "PlayerMgr", FuncName: "Remove", Uid: p.uid}
	// 建立连接
	wsUrl := fmt.Sprintf("ws://%s:%d/ws", p.cfg.Ip, p.cfg.Port)
	ws, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		actor.SendMsg(head, p.uid)
		return err
	}
	p.conn = network.NewSocket(ws, frame.New(p.node))

	// 设置 session
	tok, err := token.GenToken(&token.Token{Uid: p.uid})
	if err != nil {
		actor.SendMsg(head, p.uid)
		return err
	}

	// 发送登录请求
	buf, _ := proto.Marshal(&pb.GateLoginRequest{Token: tok})
	if err := p.SendCmd(st, uint32(pb.CMD_CMDGATE_LOGIN_REQ), p.uid, buf); err != nil {
		actor.SendMsg(head, p.uid)
		return err
	}

	// 接收登录返回消息
	pack := &pb.Packet{}
	if err := p.conn.Read(pack); err != nil {
		mlog.Fatalf("----------------------------------")
		actor.SendMsg(head, p.uid)
		return err
	}
	ms := time.Now().UnixMilli()
	flag := true
	defer p.finish(pack.Head.Cmd, ms, flag)

	loginRsp := &pb.GateLoginResponse{}
	if err := proto.Unmarshal(pack.Body, loginRsp); err != nil {
		actor.SendMsg(head, p.uid)
		return err
	}
	if loginRsp.Head != nil {
		flag = false
		return uerror.ToError(loginRsp.Head)
	}
	safe.Go(p.loop)
	safe.Go(p.keepAlive)
	mlog.Tracef("登录成功: %d", p.uid)
	return nil
}

func (p *ClientPlayer) finish(cmd uint32, ms int64, flag bool) {
	if rr := p.cmds[cmd-cmd%2].Load(); rr != nil {
		if ret := rr.Get(p.uid); ret != nil {
			ret.Finish(ms, flag)
			rr.Done()
		}
	}
}

func (p *ClientPlayer) SendCmd(st *stat.CmdStat, cmd uint32, routeId uint64, buf []byte) error {
	if rr := p.cmds[cmd].Load(); rr != nil {
		if ret := rr.Get(p.uid); ret != nil && !ret.IsFinish() {
			return uerror.New(pb.ErrorCode_REQUEST_FAIELD, "尚未接收到应答")
		}
	}
	if st != nil {
		if rr := st.Get(p.uid); rr != nil {
			rr.Start(time.Now().UnixMilli())
			st.Add(1)
			p.cmds[cmd].Store(st)
		}
	}

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

// 循环发送心跳
func (p *ClientPlayer) keepAlive() {
	tt := time.NewTicker(3 * time.Second)
	defer tt.Stop()
	buf, _ := proto.Marshal(&pb.GateHeartRequest{})
	for {
		<-tt.C
		p.SendMsg(&pb.Head{FuncName: "SendCmd"}, nil, uint32(pb.CMD_CMDGATE_HEART_REQ), p.uid, buf)
	}
}

func (p *ClientPlayer) loop() {
	for {
		pack := &pb.Packet{}
		if err := p.conn.Read(pack); err != nil {
			mlog.Errorf("读取消息失败: %v", err)
			break
		}
		ms := time.Now().UnixMilli()
		flag := true

		switch pack.Head.Cmd {
		case uint32(pb.CMD_CMDGATE_HEART_RSP):
		default:
			if ff, ok := request.Cmds[pack.Head.Cmd]; ok && pack.Head.Cmd%2 == 1 {
				msg := ff().(define.IRspProto)
				if err := proto.Unmarshal(pack.Body, msg); err != nil {
					mlog.Errorf("反序列化失败: %v", err)
					break
				}
				if msg.GetHead() != nil {
					flag = false
				}
				//mlog.Infof("[%d] [%s] rsp:%v", p.uid, pb.CMD(pack.Head.Cmd).String(), msg.GetHead())
			}
		}
		p.finish(pack.Head.Cmd, ms, flag)
	}
}
