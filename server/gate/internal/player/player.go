package player

import (
	"go-actor/common/pb"
	"go-actor/common/token"
	"go-actor/framework"
	"go-actor/framework/actor"
	"go-actor/framework/cluster"
	"go-actor/framework/define"
	"go-actor/framework/network"
	"go-actor/library/mlog"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

type Player struct {
	actor.Actor
	inet       define.INet // 网络连接
	status     int32       // 玩家登录状态
	createTime int64       // 创建时间
	extra      uint32      // 设备唯一 id
	version    uint32      // 版本号
}

func NewPlayer(conn *websocket.Conn, fr define.IFrame) *Player {
	p := &Player{}
	p.Actor.Register(p)
	p.Actor.Start()
	p.inet = network.NewSocket(conn, fr)
	return p
}

func (p *Player) Close() {
	uid := p.GetId()
	p.Actor.Stop()
	p.inet.Close()
	mlog.Infof("关闭玩家actor(%d)", uid)
}

func (p *Player) GetExtra() uint32 {
	return atomic.LoadUint32(&p.extra)
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
	p.version = pack.Head.Version
	p.extra = pack.Head.Extra
	return cluster.SendToGame(pack.Head, "PlayerDataMgr.Login", req)
}

func (p *Player) Kick(extra uint32) {
	// 不是同一设备，发送剔除消息
	if extra != p.extra {
		uid := p.GetId()
		p.SendToClient(&pb.Head{
			Src: framework.NewSrcRouter(pb.RouterType_UID, uid),
			Cmd: uint32(pb.CMD_CMDKICK_PLAYER_NTF),
			Uid: uid,
		}, &pb.KickPlayerNotify{})
	}
}

// 登录成功请求
func (p *Player) LoginSuccess(head *pb.Head, req *pb.GateLoginRequest, rsp *pb.GateLoginResponse) error {
	p.status = 1
	return p.SendToClient(head, rsp)
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
	atomic.AddUint32(&head.Reference, 1)
	mlog.Trace(head, "向客户端发送消息：%v", msg)
	return p.inet.Write(&pb.Packet{Head: head, Body: buf})
}

// 消息分发处理( 接受 websocket 传过来的消息)
func (p *Player) Dispatcher() {
	for {
		// 从客户端持续接受包消息
		pack := &pb.Packet{}
		if err := p.inet.Read(pack); err != nil {
			mlog.Errorf("读取数据包失败, websocket异常中断: %v", err)
			actor.SendMsg(&pb.Head{ActorName: "GatePlayerMgr", FuncName: "Kick", Uid: p.GetId()})
			return
		}
		mlog.Trace(pack.Head, "status(%d)收到客户端数据包: %v", p.status, pack.Body)

		// 为登录成功，任何请求直接丢弃
		if p.status <= 0 {
			continue
		}

		// 处理包消息
		switch pack.Head.Dst.NodeType {
		case pb.NodeType_NodeTypeGate: //本地调用
			pack.Head.ActorName, pack.Head.FuncName = framework.ParseActorFunc(pack.Head.Dst.ActorFunc)
			if pack.Head.Dst.ActorId <= 0 {
				pack.Head.ActorId = pack.Head.Dst.RouterId
			} else {
				pack.Head.ActorId = pack.Head.Dst.ActorId
			}
			if err := actor.Send(pack.Head, pack.Body); err != nil {
				mlog.Error(pack.Head, "gate服务Actor调用 error:%v", err)
			} else {
				mlog.Trace(pack.Head, "gate服务Actor调用 %v", pack.Body)
			}
		case pb.NodeType_NodeTypeRoom: //转发战斗服
			pack.Head.Dst.RouterType = pb.RouterType_ROOM_ID
			if err := cluster.Send(pack.Head, pack.Body); err != nil {
				mlog.Error(pack.Head, "[GateSendToRoom] 转发websocket数据包失败 error:%v", err)
			} else {
				mlog.Trace(pack.Head, "[GateSendToRoom] gate服务转发到room： %v", pack.Body)
			}
		default: //转发其他服务
			pack.Head.Dst.RouterId = pack.Head.Uid
			pack.Head.Dst.RouterType = pb.RouterType_UID
			if err := cluster.Send(pack.Head, pack.Body); err != nil {
				mlog.Error(pack.Head, "[GateSendToDefault] 转发websocket数据包失败 error:%v", err)
			} else {
				mlog.Trace(pack.Head, "[GateSendToDefault] gate服务转发到%s： %v", pack.Head.Dst.NodeType, pack.Body)
			}
		}
	}
}
