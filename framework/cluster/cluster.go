package cluster

import (
	"go-actor/common/pb"
	"go-actor/framework/define"
	"go-actor/library/encode"
	"go-actor/library/mlog"
	"go-actor/library/uerror"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
)

// 集群
type Cluster struct {
	nodeMgr      define.INode
	tableMgr     define.ITable
	discoveryObj define.IDiscovery
	busObj       define.IBus
}

func New(n define.INode, t define.ITable, b define.IBus, d define.IDiscovery) *Cluster {
	return &Cluster{n, t, d, b}
}

func (c *Cluster) Close() {
	c.tableMgr.Close()
	c.discoveryObj.Close()
	c.busObj.Close()
}

func (c *Cluster) GetSelf() *pb.Node {
	return c.nodeMgr.GetSelf()
}

func (c *Cluster) SetBroadcastHandler(f func(*pb.Head, []byte)) error {
	return c.busObj.SetBroadcastHandler(c.nodeMgr.GetSelf(), f)
}

func (c *Cluster) SetSendHandler(f func(*pb.Head, []byte)) error {
	return c.busObj.SetSendHandler(c.nodeMgr.GetSelf(), func(head *pb.Head, body []byte) {
		c.updateRouter(head.Src, head.Dst)
		if head.SendType != pb.SendType_ROUTER {
			f(head, body)
		} else {
			dd := &pb.RouterNotify{}
			proto.Unmarshal(body, dd)
			self := c.nodeMgr.GetSelf()
			for _, item := range dd.List {
				c.tableMgr.Add(item.RouterType, item.RouterId, self, item.Router)
			}
		}
	})
}

func (c *Cluster) SetReplyHandler(f func(*pb.Head, []byte)) error {
	return c.busObj.SetReplyHandler(c.nodeMgr.GetSelf(), func(head *pb.Head, body []byte) {
		c.updateRouter(head.Src, head.Dst)
		f(head, body)
	})
}

func (c *Cluster) updateRouter(rrs ...*pb.NodeRouter) {
	for _, rr := range rrs {
		if rr != nil && rr.Router != nil {
			c.tableMgr.GetOrNew(rr.RouterType, rr.RouterId, c.nodeMgr.GetSelf()).SetData(rr.Router)
			rr.Router = nil
		}
	}
}

func (c *Cluster) queryRouter(rrs ...*pb.NodeRouter) {
	for _, rr := range rrs {
		if rr != nil {
			rr.Router = c.tableMgr.GetOrNew(rr.RouterType, rr.RouterId, c.nodeMgr.GetSelf()).GetData()
		}
	}
}

func (c *Cluster) Broadcast(head *pb.Head, args ...interface{}) error {
	if head.Dst == nil {
		return uerror.New(pb.ErrorCode_PARAM_INVALID, "参数错误")
	}
	if head.Dst.NodeType <= pb.NodeType_NodeTypeBegin || head.Dst.NodeType >= pb.NodeType_NodeTypeEnd {
		return uerror.New(pb.ErrorCode_NODE_TYPE_NOT_SUPPORTED, "节点类型不支持")
	}
	if c.nodeMgr.GetCount(head.Dst.NodeType) <= 0 {
		return uerror.New(pb.ErrorCode_NODE_NOT_FOUND, "服务节点不存在")
	}
	buf, err := encode.Marshal(args...)
	if err != nil {
		return err
	}
	return c.busObj.Broadcast(head, buf)
}

func (c *Cluster) Send(head *pb.Head, args ...interface{}) error {
	if err := c.dispatcher(head); err != nil {
		return err
	}
	if head.Src == nil {
		return uerror.New(pb.ErrorCode_PARAM_INVALID, "参数错误")
	}
	c.queryRouter(head.Dst, head.Src)
	if head.Cmd > 0 && head.Cmd%2 == 0 {
		atomic.AddUint32(&head.Reference, 1)
	}
	buf, err := encode.Marshal(args...)
	if err != nil {
		return uerror.Err(pb.ErrorCode_MARSHAL_FAILED, err)
	}
	return c.busObj.Send(head, buf)
}

func (c *Cluster) Request(head *pb.Head, msg interface{}, rsp proto.Message) error {
	if err := c.dispatcher(head); err != nil {
		return err
	}
	c.queryRouter(head.Dst, head.Src)
	buf, err := encode.Marshal(msg)
	if err != nil {
		return uerror.Err(pb.ErrorCode_MARSHAL_FAILED, err)
	}
	return c.busObj.Request(head, buf, rsp)
}

func (c *Cluster) Response(head *pb.Head, msg interface{}) error {
	buf, err := encode.Marshal(msg)
	if err != nil {
		return uerror.Err(pb.ErrorCode_MARSHAL_FAILED, err)
	}
	c.queryRouter(head.Dst, head.Src)
	return c.busObj.Response(head, buf)
}

func (c *Cluster) dispatcher(head *pb.Head) error {
	if head.Dst == nil {
		return uerror.New(pb.ErrorCode_PARAM_INVALID, "参数错误")
	}
	if head.Dst.NodeType >= pb.NodeType_NodeTypeEnd || head.Dst.NodeType <= pb.NodeType_NodeTypeBegin {
		return uerror.New(pb.ErrorCode_NODE_TYPE_NOT_SUPPORTED, "节点类型不支持")
	}
	self := c.nodeMgr.GetSelf()
	if head.Dst.NodeType == self.Type {
		return uerror.New(pb.ErrorCode_NODE_TYPE_NOT_SUPPORTED, "禁止同节点类型发送")
	}
	if head.Dst.NodeId > 0 {
		if c.nodeMgr.Get(head.Dst.NodeType, head.Dst.NodeId) != nil {
			return nil
		}
		return uerror.New(pb.ErrorCode_NODE_NOT_FOUND, "节点不存在")
	}
	// 从路由表中读取
	dstTab := c.tableMgr.GetOrNew(head.Dst.RouterType, head.Dst.RouterId, self)
	if nodeId := dstTab.Get(head.Dst.NodeType); nodeId > 0 {
		if c.nodeMgr.Get(head.Dst.NodeType, nodeId) != nil {
			head.Dst.NodeId = nodeId
			return nil
		}
	}
	// 从集群节点中随机
	if nn := c.nodeMgr.Random(head.Dst.NodeType, head.Dst.RouterId); nn != nil {
		head.Dst.NodeId = nn.Id
		return nil
	}
	return uerror.New(pb.ErrorCode_NODE_NOT_FOUND, "节点不存在")
}

func (c *Cluster) SendToClient(head *pb.Head, msg proto.Message, uids ...uint64) error {
	if head.Uid > 0 {
		uids = append(uids, head.Uid)
	}
	if len(uids) <= 0 {
		return uerror.New(pb.ErrorCode_PARAM_INVALID, "参数错误")
	}
	buf, err := encode.Marshal(msg)
	if err != nil {
		return uerror.Err(pb.ErrorCode_MARSHAL_FAILED, err)
	}
	self := c.nodeMgr.GetSelf()
	if self.Type == pb.NodeType_NodeTypeGate {
		return uerror.New(pb.ErrorCode_NODE_TYPE_NOT_SUPPORTED, "禁止同节点类型发送")
	}
	atomic.AddUint32(&head.Reference, 1)
	tmps := map[uint64]struct{}{}
	for _, uid := range uids {
		if _, ok := tmps[uid]; !ok {
			tmps[uid] = struct{}{}
		} else {
			continue
		}

		dstTab := c.tableMgr.Get(pb.RouterType_UID, uid)
		if dstTab == nil {
			mlog.Warnf("玩家路由不存在 uid:%d", uid)
			continue
		}
		head.Uid = uid
		head.Dst = &pb.NodeRouter{
			NodeType:   pb.NodeType_NodeTypeGate,
			NodeId:     dstTab.Get(pb.NodeType_NodeTypeGate),
			RouterType: pb.RouterType_UID,
			RouterId:   uid,
			Router:     dstTab.GetData(),
		}
		if err := c.busObj.Send(head, buf); err != nil {
			mlog.Errorf("发送客户端失败：%v", err)
		}
	}
	return nil
}

func (c *Cluster) SendResponse(head *pb.Head, rsp proto.Message) (err error) {
	defer mlog.Trace(head, "SendToClient Rsp<%v>|Error<%v>", rsp, err)
	if len(head.Reply) > 0 {
		err = c.Response(head, rsp)
		return
	}
	if head.Cmd > 0 {
		head.Src = head.Dst
		err = c.SendToClient(head, rsp)
		return
	}
	if head.Src != nil && len(head.Src.ActorFunc) > 0 {
		head.Src, head.Dst = head.Dst, head.Src
		err = c.Send(head, rsp)
	}
	return
}

// 同步路由
func (c *Cluster) SendRouter(head *pb.Head, rt pb.RouterType, ids ...uint64) error {
	if head.Dst == nil {
		return uerror.New(pb.ErrorCode_PARAM_INVALID, "参数错误")
	}
	if head.Dst.NodeType >= pb.NodeType_NodeTypeEnd || head.Dst.NodeType <= pb.NodeType_NodeTypeBegin {
		return uerror.New(pb.ErrorCode_NODE_TYPE_NOT_SUPPORTED, "节点类型不支持")
	}
	self := c.nodeMgr.GetSelf()
	if head.Dst.NodeType == self.Type {
		return uerror.New(pb.ErrorCode_NODE_TYPE_NOT_SUPPORTED, "禁止同节点类型发送")
	}
	// 从路由表中读取
	dstTab := c.tableMgr.Get(head.Dst.RouterType, head.Dst.RouterId)
	if dstTab == nil {
		return uerror.New(pb.ErrorCode_PARAM_INVALID, "路由不存在，无法更新路由")
	}
	head.SendType = pb.SendType_ROUTER
	head.Dst.NodeId = dstTab.Get(head.Dst.NodeType)
	head.Dst.Router = dstTab.GetData()
	if c.nodeMgr.Get(head.Dst.NodeType, head.Dst.NodeId) == nil {
		return uerror.New(pb.ErrorCode_NODE_NOT_FOUND, "节点不存在")
	}
	// 发送通知
	event := &pb.RouterNotify{}
	for _, id := range ids {
		if rr := c.tableMgr.Get(rt, id); rr != nil {
			event.List = append(event.List, &pb.NodeRouter{
				RouterType: rt,
				RouterId:   id,
				Router:     rr.GetData(),
			})
		}
	}
	buf, err := encode.Marshal(event)
	if err != nil {
		return uerror.Err(pb.ErrorCode_MARSHAL_FAILED, err)
	}
	return c.busObj.Send(head, buf)
}
