package service

import (
	"poker_server/common/pb"
	"poker_server/common/yaml"
	"poker_server/framework/domain"
	"poker_server/framework/internal/core/bus"
	"poker_server/framework/internal/core/cluster"
	"poker_server/framework/internal/core/discovery"
	"poker_server/framework/internal/core/router"
	"poker_server/library/encode"
	"poker_server/library/mlog"
	"poker_server/library/uerror"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
)

type Service struct {
	node         *pb.Node          // 节点信息
	clusterObj   domain.ICluster   // 集群节点
	tableObj     domain.ITable     // 路由表
	discoveryObj domain.IDiscovery // 服务发现
	busObj       domain.IBus       // 消息总线
}

func NewService(node *pb.Node, server *yaml.ServerConfig, cfg *yaml.Config) (*Service, error) {
	clusterObj := cluster.New()
	tableObj := router.New()
	tableObj.SetExpire(cfg.Common.RouterExpire)

	// 服务发现
	dis, err := discovery.NewEtcd(cfg.Etcd)
	if err != nil {
		return nil, err
	}
	if err := dis.Watch(clusterObj); err != nil {
		return nil, err
	}
	if err := dis.Register(node, cfg.Common.DiscoveryExpire); err != nil {
		return nil, err
	}

	// 消息中间件
	busObj, err := bus.NewNats(cfg.Nats, tableObj)
	if err != nil {
		return nil, err
	}
	return &Service{
		node:         node,
		clusterObj:   clusterObj,
		tableObj:     tableObj,
		discoveryObj: dis,
		busObj:       busObj,
	}, nil
}

func (d *Service) Close() error {
	return d.discoveryObj.Close()
}

func (d *Service) GetNode() *pb.Node {
	return d.node
}

func (d *Service) RegisterBroadcastHandler(f func(*pb.Head, []byte)) error {
	return d.busObj.SetBroadcastHandler(d.node, f)
}

func (d *Service) RegisterSendHandler(f func(*pb.Head, []byte)) error {
	return d.busObj.SetSendHandler(d.node, f)
}

func (d *Service) RegisterReplyHandler(f func(*pb.Head, []byte)) error {
	return d.busObj.SetReplyHandler(d.node, f)
}

func (d *Service) Broadcast(head *pb.Head, args ...interface{}) error {
	// 检测参数
	if err := d.checkDst(head); err != nil {
		return err
	}

	// 设置值
	head.SendType = pb.SendType_BROADCAST
	if head.Src != nil {
		d.checkSrc(head)
	} else {
		head.Src = &pb.NodeRouter{NodeType: d.node.Type, NodeId: d.node.Id}
	}

	// 解析参数
	buf, err := parseArgs(args...)
	if err != nil {
		return err
	}
	return d.busObj.Broadcast(head, buf)
}

func (d *Service) Send(head *pb.Head, args ...interface{}) error {
	// 检测参数
	if err := d.checkSrc(head); err != nil {
		return err
	}
	if err := d.checkDst(head); err != nil {
		return err
	}

	// 做路由分发
	if err := d.dispatcher(head); err != nil {
		return err
	}

	// 检测参数
	head.SendType = pb.SendType_POINT
	if head.Dst.NodeType == d.node.Type && head.Dst.NodeId == d.node.Id {
		return uerror.New(1, pb.ErrorCode_REQUEST_FAIELD, "不能发送给自身节点: %v", head)
	}

	// 解析参数
	atomic.AddInt32(&head.Reference, 1)
	buf, err := parseArgs(args...)
	if err != nil {
		return err
	}
	return d.busObj.Send(head, buf)
}

func (d *Service) Request(head *pb.Head, msg interface{}, reply proto.Message) error {
	// 检测参数
	if err := d.checkDst(head); err != nil {
		return err
	}
	if err := d.dispatcher(head); err != nil {
		return err
	}

	// 做路由分发
	if head.Src != nil {
		d.checkSrc(head)
	} else {
		head.Src = &pb.NodeRouter{NodeType: d.node.Type, NodeId: d.node.Id}
	}

	// 检测参数
	head.SendType = pb.SendType_POINT
	if head.Dst.NodeType == d.node.Type && head.Dst.NodeId == d.node.Id {
		return uerror.New(1, pb.ErrorCode_REQUEST_FAIELD, "不能发送给自身节点: %v", head)
	}

	// 解析参数
	buf, err := parseArgs(msg)
	if err != nil {
		return err
	}
	return d.busObj.Request(head, buf, reply)
}

func (d *Service) Response(head *pb.Head, msg interface{}) error {
	if len(head.Reply) <= 0 {
		return nil
	}
	head.SendType = pb.SendType_POINT

	// 解析参数
	buf, err := parseArgs(msg)
	if err != nil {
		return err
	}
	return d.busObj.Response(head, buf)
}

func (d *Service) SendToClient(head *pb.Head, msg proto.Message) error {
	// 检测参数
	if err := d.checkSrc(head); err != nil {
		return err
	}
	if head.Dst == nil && head.Uid <= 0 {
		return uerror.New(1, pb.ErrorCode_PARAM_INVALID, "玩家UID为空: %v", head)
	}

	// 解析参数
	buf, err := proto.Marshal(msg)
	if err != nil {
		return uerror.New(1, pb.ErrorCode_MARSHAL_FAILED, "序列化失败：%v", err)
	}
	return d.sendToClient(head, buf)
}

func (d *Service) NotifyToClient(uids []uint64, head *pb.Head, msg proto.Message) error {
	// 检测参数
	if len(uids) <= 0 {
		return nil
	}
	if err := d.checkSrc(head); err != nil {
		return err
	}

	// 序列化数据
	buf, err := proto.Marshal(msg)
	if err != nil {
		return uerror.New(1, pb.ErrorCode_MARSHAL_FAILED, "序列化失败：%v", err)
	}
	for _, uid := range uids {
		head.Uid = uid
		if err := d.sendToClient(head, buf); err != nil {
			mlog.Errorf("通知玩家失败：%v, error:%v", head, err)
		}
	}
	return nil
}

func (d *Service) dispatcher(head *pb.Head) error {
	route := d.tableObj.Get(head.Dst.RouterType, head.Dst.ActorId)
	route.Set(d.node.Type, d.node.Id)
	head.Dst.Router = route.GetData()

	// 业务层直接指定具体节点
	if head.Dst.NodeId > 0 {
		if d.clusterObj.Get(head.Dst.NodeType, head.Dst.NodeId) != nil {
			route.Set(head.Dst.NodeType, head.Dst.NodeId)
			return nil
		}
		return uerror.New(1, pb.ErrorCode_NODE_NOT_FOUND, "未找到服务节点: %v", head)
	}

	// 优先从路由中选择
	if nodeId := route.Get(head.Dst.NodeType); nodeId > 0 {
		if d.clusterObj.Get(head.Dst.NodeType, nodeId) != nil {
			route.Set(head.Dst.NodeType, nodeId)
			head.Dst.NodeId = nodeId
			return nil
		}
	}

	//从集群中随机获取一个节点
	if node := d.clusterObj.Random(head.Dst.NodeType, head.Dst.ActorId); node != nil {
		route.Set(head.Dst.NodeType, node.Id)
		head.Dst.NodeId = node.Id
		return nil
	}
	return uerror.New(1, pb.ErrorCode_NODE_NOT_FOUND, "未找到服务节点: %v", head)
}

func (d *Service) sendToClient(head *pb.Head, buf []byte) error {
	head.Dst = &pb.NodeRouter{
		NodeType: pb.NodeType_NodeTypeGate,
		//	NodeId:     route.Get(pb.NodeType_NodeTypeGate),
		RouterType: pb.RouterType_RouterTypeUid,
		ActorName:  "Player",
		ActorId:    head.Uid,
		//	Router:     route.GetData(),
	}
	// 从路由中读取信息
	route := d.tableObj.Get(head.Dst.RouterType, head.Dst.ActorId)
	head.Dst.NodeId = route.Get(head.Dst.NodeType)
	head.Dst.Router = route.GetData()

	// 判断节点是否还在
	if d.clusterObj.Get(head.Dst.NodeType, head.Dst.NodeId) == nil {
		return uerror.New(1, pb.ErrorCode_NODE_NOT_FOUND, "未找到服务节点: %v", head)
	}

	// 设置发送类型
	if head.Cmd%2 == 0 {
		if _, ok := pb.CMD_name[int32(head.Cmd)+1]; ok {
			head.Cmd++
			head.Seq++
		}
	}
	mlog.Debug(head, "发送消息到客户端: %d -> %v", len(buf), buf)
	return d.busObj.Send(head, buf)
}

func (d *Service) checkDst(head *pb.Head) error {
	// 检测参数
	if head.Dst == nil || head.Dst.RouterType <= pb.RouterType_RouterTypeNone || head.Dst.ActorId <= 0 {
		return uerror.New(1, pb.ErrorCode_PARAM_INVALID, "head.Dst为空: %v", head)
	}

	// 判断类型是否支持
	if head.Dst.NodeType >= pb.NodeType_NodeTypeEnd || head.Dst.NodeType <= pb.NodeType_NodeTypeBegin {
		return uerror.New(1, pb.ErrorCode_PARAM_INVALID, "服务类型不支持: %v", head)
	}

	// 判断节点是否存在
	if d.clusterObj.GetCount(head.Dst.NodeType) <= 0 {
		return uerror.New(1, pb.ErrorCode_NODE_NOT_FOUND, "未找到服务节点: %v", head)
	}
	return nil
}

func (d *Service) checkSrc(head *pb.Head) error {
	if head.Src != nil {
		head.Src.NodeType = d.node.Type
		head.Src.NodeId = d.node.Id
	}

	// 检测参数
	if head.Src == nil || head.Src.RouterType <= pb.RouterType_RouterTypeNone || head.Src.ActorId <= 0 {
		return uerror.New(1, pb.ErrorCode_PARAM_INVALID, "head.Src为空: %v", head)
	}

	// 设置自身路由
	route := d.tableObj.Get(head.Src.RouterType, head.Src.ActorId)
	route.Set(d.node.Type, d.node.Id)
	head.Src.Router = route.GetData()
	return nil
}

func parseArgs(args ...interface{}) ([]byte, error) {
	if len(args) == 1 {
		switch vv := args[0].(type) {
		case []byte:
			return vv, nil
		case proto.Message:
			buf, err := proto.Marshal(vv)
			if err != nil {
				return nil, uerror.New(1, pb.ErrorCode_MARSHAL_FAILED, "序列化失败：%v", err)
			}
			return buf, nil
		}
	}
	return encode.Encode(args...), nil
}
