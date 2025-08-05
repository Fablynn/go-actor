package bus

import (
	"fmt"
	"go-actor/common/pb"
	"go-actor/common/yaml"
	"go-actor/framework/define"
	"go-actor/library/mlog"
	"go-actor/library/util"
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
	client *nats.Conn
	topic  string
	table  define.ITable
}

func NewNats(cfg *yaml.NatsConfig, tab define.ITable) (cli *Nats, err error) {
	err = util.Retry(3, time.Second, func() error {
		client, err := nats.Connect(cfg.Endpoints, close, waitRecon, reconErr, disconErr)
		if err == nil {
			cli = &Nats{
				topic:  cfg.Topic,
				client: client,
				table:  tab,
			}
		}
		return err
	})
	return
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
	_, err := n.client.Subscribe(n.broadcastChannel(node.Type), func(msg *nats.Msg) {
		pack := pb.Packet{}
		if err := proto.Unmarshal(msg.Data, &pack); err != nil {
			mlog.Errorf("nats解析packet包失败Error<%v>, buf:%v", err, msg.Data)
			return
		}
		mlog.Tracef("nats接收到数据包 %v", pack)

		ff(pack.Head, pack.Body)
	})
	return err
}

// 单播
func (n *Nats) SetSendHandler(node *pb.Node, ff func(*pb.Head, []byte)) error {
	_, err := n.client.Subscribe(n.sendChannel(node.Type, node.Id), func(msg *nats.Msg) {
		pack := pb.Packet{}
		if err := proto.Unmarshal(msg.Data, &pack); err != nil {
			mlog.Errorf("nats解析packet包失败Error<%v>, buf:%v", err, msg.Data)
			return
		}
		mlog.Tracef("nats接收到数据包 %v", pack)

		ff(pack.Head, pack.Body)
	})
	return err
}

// rpc调用
func (n *Nats) SetReplyHandler(node *pb.Node, ff func(*pb.Head, []byte)) error {
	_, err := n.client.Subscribe(n.replyChannel(node.Type, node.Id), func(msg *nats.Msg) {
		pack := pb.Packet{}
		if err := proto.Unmarshal(msg.Data, &pack); err != nil {
			mlog.Errorf("nats解析packet包失败Error<%v>, buf:%v", err, msg.Data)
			return
		}
		pack.Head.Reply = msg.Reply
		mlog.Tracef("nats接收到数据包 %v", pack)

		ff(pack.Head, pack.Body)
	})
	return err
}

// 发送广播
func (d *Nats) Broadcast(head *pb.Head, msg []byte) error {
	msgBuf, err := proto.Marshal(&pb.Packet{
		Head: &pb.Head{
			SendType: pb.SendType_BROADCAST,
			Src:      head.Src,
			Dst:      head.Dst,
			Uid:      head.Uid,
			Seq:      head.Seq,
			Cmd:      head.Cmd,
			Reply:    head.Reply,
		},
		Body: msg,
	})
	if err != nil {
		return err
	}
	return d.client.Publish(d.broadcastChannel(head.Dst.NodeType), msgBuf)
}

// 发送请求
func (d *Nats) Send(head *pb.Head, msg []byte) error {
	msgBuf, err := proto.Marshal(&pb.Packet{
		Head: &pb.Head{
			Src:   head.Src,
			Dst:   head.Dst,
			Uid:   head.Uid,
			Seq:   head.Seq,
			Cmd:   head.Cmd,
			Reply: head.Reply,
		},
		Body: msg,
	})
	if err != nil {
		return err
	}
	return d.client.Publish(d.sendChannel(head.Dst.NodeType, head.Dst.NodeId), msgBuf)
}

// 发送同步请求
func (d *Nats) Request(head *pb.Head, req []byte, rsp proto.Message) error {
	msgBuf, err := proto.Marshal(&pb.Packet{
		Head: &pb.Head{
			Src:   head.Src,
			Dst:   head.Dst,
			Uid:   head.Uid,
			Seq:   head.Seq,
			Cmd:   head.Cmd,
			Reply: head.Reply,
		},
		Body: req,
	})
	if err != nil {
		return err
	}
	resp, err := d.client.Request(d.replyChannel(head.Dst.NodeType, head.Dst.NodeId), msgBuf, 3*time.Second)
	if err != nil {
		return err
	}
	return proto.Unmarshal(resp.Data, rsp)
}

// 发送同步应答
func (d *Nats) Response(head *pb.Head, msg []byte) error {
	return d.client.Publish(head.Reply, msg)
}

func (n *Nats) Close() {
	if n.client != nil {
		n.client.Close()
	}
}
