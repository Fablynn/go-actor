package player

import (
	"go-actor/common/pb"
	"go-actor/framework"
	"go-actor/framework/actor"
	"go-actor/framework/domain"
	"go-actor/framework/network"
	"go-actor/framework/token"
	"go-actor/library/mlog"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

type Player struct {
	actor.Actor
	inet       domain.INet // 网络连接
	status     int32       // 玩家登录状态
	createTime int64       // 创建时间
	// 加密方式 todo
}

func NewPlayer(conn *websocket.Conn, fr domain.IFrame) *Player {
	p := &Player{}
	p.Actor.Register(p)
	p.inet = network.NewSocket(conn, 1024*1024)
	p.inet.Register(fr)
	return p
}

func (p *Player) CheckToken() error {
	// 第一个包一定是登录认证包
	p.inet.SetReadExpire(5)
	pack := &pb.Packet{}
	if err := p.inet.Read(pack); err != nil {
		return err
	}
	p.inet.SetReadExpire(0)
	req := &pb.GateLoginRequest{}
	if err := proto.Unmarshal(pack.Body, req); err != nil {
		return err
	}

	// 解析token
	tt, err := token.ParseToken(req.Token)
	if err != nil {
		return err
	}

	// 设置玩家ID
	pack.Head.Uid = tt.Uid
	now := time.Now().Unix()
	p.Actor.SetId(tt.Uid)
	p.createTime = now

	head := framework.SwapToGame(pack.Head, tt.Uid, "PlayerDataMgr", "Login")
	head.Src = framework.NewSrcRouter(pb.RouterType_RouterTypeUid, tt.Uid)
	head.Dst.RouterType = pb.RouterType_RouterTypeUid
	return framework.Send(head, req)
}

// 登录成功请求
func (p *Player) LoginSuccess(head *pb.Head, req *pb.GateLoginRequest, rsp *pb.GateLoginResponse) error {
	p.status = 1
	return p.SendToClient(head, rsp)
}

func (p *Player) Start() {
	p.Actor.Start()
}

func (p *Player) Stop() {
	p.inet.Close()
	p.Actor.Stop()
}

// 向客户端发送数据
func (p *Player) SendToClient(head *pb.Head, msg interface{}) error {
	var buf []byte
	switch vv := msg.(type) {
	case []byte:
		buf = vv
	case proto.Message:
		buf, _ = proto.Marshal(vv)
	}
	if head.Cmd%2 == 0 {
		if _, ok := pb.CMD_name[int32(head.Cmd)+1]; ok {
			head.Cmd++
			head.Seq++
		}
	}
	atomic.AddInt32(&head.Reference, 1)
	return p.inet.Write(&pb.Packet{Head: head, Body: buf})
}

// 消息分发处理( 接受 websocket 传过来的消息)
func (p *Player) Dispatcher() {
	for {
		// 从客户端持续接受包消息
		pack := &pb.Packet{}
		if err := p.inet.Read(pack); err != nil {
			mlog.Errorf("读取数据包失败, websocket异常中断: %v", err)
			return
		}

		// 为登录成功，任何请求直接丢弃
		if p.status <= 0 {
			continue
		}

		// 处理包消息
		switch pack.Head.Dst.NodeType {
		case pb.NodeType_NodeTypeGate:
			pack.Head.ActorName = pack.Head.Dst.ActorName
			pack.Head.FuncName = pack.Head.Dst.FuncName
			pack.Head.ActorId = pack.Head.Dst.ActorId
			mlog.Debug(pack.Head, "收到websocket数据包 pack:%v", pack)

			// gate直接处理
			if err := actor.Send(pack.Head, pack.Body); err != nil {
				mlog.Error(pack.Head, "gate服务Actor调用: %v", err)
			}

		case pb.NodeType_NodeTypeGame:
			pack.Head.Dst.ActorId = pack.Head.Uid
			pack.Head.Src = framework.NewSrcRouter(pb.RouterType_RouterTypeUid, p.GetId(), "Dispatcher")
			if err := framework.Send(pack.Head, pack.Body); err != nil {
				mlog.Error(pack.Head, "转发websocket数据包失败: pack:%v, error:%v", pack, err)
			} else {
				mlog.Debug(pack.Head, "转发websocket数据包 pack:%v", pack)
			}
		default:
			// 转发到其他服务
			pack.Head.Src = framework.NewSrcRouter(pb.RouterType_RouterTypeUid, p.GetId(), "Dispatcher")
			if err := framework.Send(pack.Head, pack.Body); err != nil {
				mlog.Error(pack.Head, "转发websocket数据包失败: pack:%v, error:%v", pack, err)
			} else {
				mlog.Debug(pack.Head, "转发websocket数据包 pack:%v", pack)
			}
		}
	}
}
