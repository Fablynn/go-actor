package define

import (
	"go-actor/common/pb"
	"time"

	"github.com/golang/protobuf/proto"
)

type IRouter interface {
	GetData() *pb.Router
	SetData(*pb.Router) IRouter
	Get(pb.NodeType) int32
	Set(pb.NodeType, int32) IRouter
}

// 路由表接口
type ITable interface {
	GetOrNew(pb.RouterType, uint64, *pb.Node) IRouter
	Get(pb.RouterType, uint64) IRouter
	Add(pb.RouterType, uint64, *pb.Node, *pb.Router)
	Close()
}

// 集群节点接口
type INode interface {
	GetSelf() *pb.Node
	GetCount(pb.NodeType) int
	Get(pb.NodeType, int32) *pb.Node
	Del(pb.NodeType, int32) bool
	Add(*pb.Node) bool
	Random(pb.NodeType, uint64) *pb.Node
}

// 服务注册与发现接口
type IDiscovery interface {
	Register(INode, int64) error
	Watch(INode) error
	Close() error
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

// 异步协程接口
type IAsync interface {
	GetIdPointer() *uint64
	GetId() uint64
	SetId(uint64)
	Start()
	Stop() // 业务层必须重载该方法
}

// 业务层必须实现的接口
type IActor interface {
	IAsync
	GetActorName() string
	Register(IActor, ...int)
	ParseFunc(interface{})
	SendMsg(*pb.Head, ...interface{}) error
	Send(*pb.Head, []byte) error
	RegisterTimer(*pb.Head, time.Duration, int32) error
}

// 数据帧接口
type IFrame interface {
	GetSize(*pb.Packet) int          // 获取包头大小
	Decode([]byte, *pb.Packet) error // 解码数据包
	Encode(*pb.Packet, []byte) error // 编码数据包
}

// (前后端)网络接口
type INet interface {
	SetReadExpire(int64)    // 设置读超时
	SetWriteExpire(int64)   // 设置写超时
	Write(*pb.Packet) error // 发送数据包
	Read(*pb.Packet) error  // 读取数据包
	Close() error           // 关闭网络
}

// 业务层Rsp必须实现的接口
type IRspProto interface {
	proto.Message
	GetHead() *pb.RspHead
	SetHead(*pb.RspHead)
}
