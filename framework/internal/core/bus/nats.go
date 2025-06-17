package bus

import (
	"fmt"
	"poker_server/common/pb"
	"poker_server/common/yaml"
	"poker_server/framework/domain"
	"poker_server/library/mlog"
	"poker_server/library/uerror"
	"poker_server/library/util"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
)

var (
	disconErr = nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
		mlog.Errorf("nats disconnect: %v", err)
	})
	reconErr = nats.ReconnectErrHandler(func(_ *nats.Conn, err error) {
		mlog.Errorf("nats reconnect: %v", err)
	})
	waitRecon = nats.ReconnectWait(5)
	close     = nats.ClosedHandler(func(_ *nats.Conn) {
		mlog.Infof("nats close")
	})
)

type Nats struct {
	topic  string
	client *nats.Conn
	table  domain.ITable
}

func NewNats(cfg *yaml.NatsConfig, table domain.ITable) (*Nats, error) {
	var client *nats.Conn
	if err := util.Retry(3, time.Second, func() error {
		cli, err := nats.Connect(cfg.Endpoints, close, waitRecon, reconErr, disconErr)
		if err == nil {
			client = cli
		}
		return err
	}); err != nil {
		return nil, uerror.New(1, pb.ErrorCode_CONNECT_FAILED, "nats服务连接失败: %v", err)
	}
	return &Nats{client: client, topic: cfg.Topic, table: table}, nil
}

func (n *Nats) broadcastChannel(t pb.NodeType) string {
	return fmt.Sprintf("%s/%d", n.topic, t)
}

func (n *Nats) sendChannel(t pb.NodeType, id int32) string {
	return fmt.Sprintf("%s/%d/%d", n.topic, t, id)
}

func (n *Nats) replyChannel(t pb.NodeType, id int32) string {
	return fmt.Sprintf("%s/rpc/%d/%d", n.topic, t, id)
}

// 广播
func (n *Nats) SetBroadcastHandler(node *pb.Node, ff func(*pb.Head, []byte)) error {
	if _, err := n.client.Subscribe(n.broadcastChannel(node.Type), func(msg *nats.Msg) {
		// 接受 nats消息
		pack := &pb.Packet{}
		if err := proto.Unmarshal(msg.Data, pack); err != nil {
			mlog.Errorf("nats解析packet包失败: %v", err)
			return
		}
		mlog.Debugf("收到Nats广播数据包 bodySize:%d, pack:%v", len(msg.Data), pack)

		// 执行函数
		ff(pack.Head, pack.Body)
	}); err != nil {
		return uerror.New(1, pb.ErrorCode_SUBSCRIBE_FAILED, "nats订阅广播失败: %v", err)
	}
	return nil
}

// 单播
func (n *Nats) SetSendHandler(node *pb.Node, ff func(*pb.Head, []byte)) error {
	if _, err := n.client.Subscribe(n.sendChannel(node.Type, node.Id), func(msg *nats.Msg) {
		// 接受 nats消息
		pack := &pb.Packet{}
		if err := proto.Unmarshal(msg.Data, pack); err != nil {
			mlog.Errorf("nats解析packet包失败: %v", err)
			return
		}
		mlog.Debugf("收到Nats单播数据包 bodySize:%d, pack:%v", len(msg.Data), pack)

		// 更新路由信息
		tabSrc := n.table.Get(pack.Head.Src.RouterType, pack.Head.Src.ActorId)
		tabSrc.SetData(pack.Head.Src.Router)
		tabSrc.Set(node.Type, node.Id)
		tabDst := n.table.Get(pack.Head.Dst.RouterType, pack.Head.Dst.ActorId)
		tabDst.SetData(pack.Head.Dst.Router)
		tabDst.Set(pack.Head.Src.NodeType, pack.Head.Src.NodeId)

		// 执行函数
		ff(pack.Head, pack.Body)
	}); err != nil {
		return uerror.New(1, pb.ErrorCode_SUBSCRIBE_FAILED, "nats订阅单播失败: %v", err)
	}
	return nil
}

// rpc调用
func (n *Nats) SetReplyHandler(node *pb.Node, ff func(*pb.Head, []byte)) error {
	if _, err := n.client.Subscribe(n.replyChannel(node.Type, node.Id), func(msg *nats.Msg) {
		// 接受 nats消息
		pack := &pb.Packet{}
		if err := proto.Unmarshal(msg.Data, pack); err != nil {
			mlog.Errorf("nats解析packet包失败: %v", err)
			return
		}
		mlog.Debugf("收到Nats广播数据包 bodySize:%d, pack:%v", len(msg.Data), pack)

		// 更新路由信息
		tabSrc := n.table.Get(pack.Head.Src.RouterType, pack.Head.Src.ActorId)
		tabSrc.SetData(pack.Head.Src.Router)
		tabSrc.Set(node.Type, node.Id)
		tabDst := n.table.Get(pack.Head.Dst.RouterType, pack.Head.Dst.ActorId)
		tabDst.SetData(pack.Head.Dst.Router)
		tabDst.Set(pack.Head.Src.NodeType, pack.Head.Src.NodeId)

		// 执行函数
		pack.Head.Reply = msg.Reply
		ff(pack.Head, pack.Body)
	}); err != nil {
		return uerror.New(1, pb.ErrorCode_SUBSCRIBE_FAILED, "nats订阅rpc调用失败: %v", err)
	}
	return nil
}

// 发送广播
func (d *Nats) Broadcast(head *pb.Head, msg []byte) error {
	reset(head, pb.SendType_BROADCAST)
	msgBuf, err := proto.Marshal(&pb.Packet{Head: head, Body: msg})
	if err != nil {
		return uerror.New(1, pb.ErrorCode_MARSHAL_FAILED, "nats待发送的msg序列化失败: %v", err)
	}
	return d.client.Publish(d.broadcastChannel(head.Dst.NodeType), msgBuf)
}

// 发送请求
func (d *Nats) Send(head *pb.Head, msg []byte) error {
	reset(head, pb.SendType_POINT)
	msgBuf, err := proto.Marshal(&pb.Packet{Head: head, Body: msg})
	if err != nil {
		return uerror.New(1, pb.ErrorCode_MARSHAL_FAILED, "nats待发送的msg序列化失败: %v", err)
	}
	return d.client.Publish(d.sendChannel(head.Dst.NodeType, head.Dst.NodeId), msgBuf)
}

// 发送同步请求
func (d *Nats) Request(head *pb.Head, req []byte, rsp proto.Message) error {
	reset(head, pb.SendType_POINT)
	msgBuf, err := proto.Marshal(&pb.Packet{Head: head, Body: req})
	if err != nil {
		return uerror.New(1, pb.ErrorCode_MARSHAL_FAILED, "nats.Request序列化失败: %v", err)
	}

	// 发送不同请求
	resp, err := d.client.Request(d.replyChannel(head.Dst.NodeType, head.Dst.NodeId), msgBuf, 3000*time.Millisecond)
	if err != nil {
		return uerror.New(1, pb.ErrorCode_REQUEST_FAIELD, "nats.Request发送请求失败: %v", err)
	}

	if err := proto.Unmarshal(resp.Data, rsp); err != nil {
		return uerror.New(1, pb.ErrorCode_UNMARSHAL_FAILED, "nats.Request解析返回值失败: %v", err)
	}
	return nil
}

// 发送同步应答
func (d *Nats) Response(head *pb.Head, msg []byte) error {
	reset(head, pb.SendType_POINT)
	return d.client.Publish(head.Reply, msg)
}

func reset(head *pb.Head, sendType pb.SendType) {
	head.ActorId = 0
	head.ActorName = ""
	head.FuncName = ""
	head.SendType = sendType
}

func (n *Nats) Close() {
	if n.client != nil {
		n.client.Close()
	}
}
