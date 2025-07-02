package manager

import (
	"go-actor/common/pb"
	"go-actor/common/yaml"
	"go-actor/framework"
	"go-actor/framework/actor"
	"go-actor/framework/token"
	"go-actor/library/async"
	"go-actor/library/mlog"
	"go-actor/server/gate/internal/http"
)

var (
	playerMgr = new(GatePlayerMgr)
)

func Init(cfg *yaml.ServerConfig, com *yaml.CommonConfig) error {
	framework.RegisterBroadcastHandler(broadcastHandler)
	framework.RegisterSendHandler(sendHandler)
	framework.RegisterReplyHandler(sendHandler)

	async.Init(mlog.Errorf)
	token.Init(com.SecretKey)

	// 初始化Actor
	async.SafeGo(mlog.Errorf, func() {
		http.Init(cfg)
	})

	// 初始化模块
	return playerMgr.Init(cfg.Ip, cfg.Port)
}

func Close() {
	playerMgr.Stop()
}

// 处理返回客户端的消息
func sendHandler(head *pb.Head, body []byte) {
	mlog.Debug(head, "收到Nats数据包 head:%v, body:%d", head, len(body))

	if len(head.Dst.FuncName) <= 0 || head.Dst.FuncName == "SendToClient" {
		head.SendType = pb.SendType_POINT
		head.Dst.ActorName = "Player"
		head.Dst.FuncName = "SendToClient"
	}

	head.ActorName = head.Dst.ActorName
	head.FuncName = head.Dst.FuncName
	head.ActorId = head.Dst.ActorId
	head.Uid = head.Dst.ActorId
	if err := actor.Send(head, body); err != nil {
		mlog.Errorf("Actor消息转发失败: %v", err)
	}
}

// 处理返回客户端的消息
func broadcastHandler(head *pb.Head, body []byte) {
	mlog.Debugf("收到Nats broadcast数据包 head:%v, body:%d", head, len(body))

	if head.Dst.FuncName == "SendToClient" || len(head.Dst.FuncName) <= 0 {
		head.SendType = pb.SendType_BROADCAST
		head.Dst.ActorName = "Player"
		head.Dst.FuncName = "SendToClient"
	}

	head.ActorName = head.Dst.ActorName
	head.FuncName = head.Dst.FuncName
	head.ActorId = head.Dst.ActorId
	head.Uid = head.Dst.ActorId
	if err := actor.Send(head, body); err != nil {
		mlog.Errorf("Actor消息转发失败: %v", err)
	}
}
