package domain

import (
	"poker_server/common/pb"
	"time"

	"github.com/golang/protobuf/proto"
)

type IRspProto interface {
	proto.Message         // 响应协议接口
	GetHead() *pb.RspHead // 获取响应头
	SetHead(*pb.RspHead)
}

// Actor接口定义
type IActor interface {
	GetId() uint64                                      // 获取Actor ID
	SetId(uint64)                                       // 设置Actor ID
	Start()                                             // 启动Actor
	Stop()                                              // 停止Actor
	GetActorName() string                               // 获取Actor名称
	Register(IActor, ...int)                            // 注册Actor
	ParseFunc(interface{})                              // 解析方法列表
	SendMsg(*pb.Head, ...interface{}) error             // 发送消息
	Send(*pb.Head, []byte) error                        // 发送远程调用
	RegisterTimer(*pb.Head, time.Duration, int32) error // 注册定时器
}

// 路由接口
type IRouter interface {
	Get(pb.NodeType) int32        // 获取路由信息
	Set(pb.NodeType, int32)       // 设置路由信息
	GetData() *pb.Router          // 获取路由信息
	SetData(*pb.Router)           // 设置路由信息
	IsExpire(now, ttl int64) bool // 是否过期
}

// 路由表接口
type ITable interface {
	Get(pb.RouterType, uint64) IRouter // 获取路由信息
	SetExpire(int64)                   // 设置路由过期时间
	Close()                            // 关闭路由管理
}

// 服务集群接口
type ICluster interface {
	GetCount(pb.NodeType) int            // 获取节点数量
	Get(pb.NodeType, int32) *pb.Node     // 获取节点
	Del(pb.NodeType, int32) bool         // 删除节点
	Add(*pb.Node) bool                   // 添加节点
	List(pb.NodeType) []*pb.Node         // 获取节点列表
	Random(pb.NodeType, uint64) *pb.Node // 随机节点
}

// 服务注册与发现接口
type IDiscovery interface {
	Register(node *pb.Node, ttl int64) error // 注册服务
	Watch(ICluster) error                    // 监听服务
	Close() error                            // 关闭服务发现
}

// 消息总线接口
type IBus interface {
	SetBroadcastHandler(*pb.Node, func(*pb.Head, []byte)) error // 监听消息
	SetSendHandler(*pb.Node, func(*pb.Head, []byte)) error      // 监听消息
	SetReplyHandler(*pb.Node, func(*pb.Head, []byte)) error     // 监听消息
	Broadcast(*pb.Head, []byte) error                           // 广播消息
	Send(*pb.Head, []byte) error                                // 发送消息
	Request(*pb.Head, []byte, proto.Message) error              // 请求应答消息
	Response(*pb.Head, []byte) error                            // 响应消息
	Close()                                                     // 关闭网络
}

// 数据帧接口
type IFrame interface {
	GetSize(*pb.Packet) int          // 获取包头大小
	Decode([]byte, *pb.Packet) error // 解码数据包
	Encode(*pb.Packet, []byte) error // 编码数据包
}

// (前后端)网络接口
type INet interface {
	Register(IFrame)        // 设置数据帧
	SetReadExpire(int64)    // 设置读超时
	SetWriteExpire(int64)   // 设置写超时
	Write(*pb.Packet) error // 发送数据包
	Read(*pb.Packet) error  // 读取数据包
	Close() error           // 关闭网络
}
