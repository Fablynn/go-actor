package internal

import (
	"fmt"
	"go-actor/common/pb"
	"go-actor/common/token"
	"go-actor/common/yaml"
	"go-actor/framework"
	"go-actor/framework/actor"
	"go-actor/framework/cluster"
	"go-actor/library/mlog"
	"go-actor/library/safe"
	"go-actor/library/util"
	"go-actor/server/gate/internal/http_api"
	"go-actor/server/gate/internal/player"
	"net/http"
)

var (
	playerMgr = new(player.GatePlayerMgr)
)

func Init(cfg *yaml.NodeConfig, com *yaml.CommonConfig) {
	util.Must(cluster.SetBroadcastHandler(handler))
	util.Must(cluster.SetSendHandler(handler))
	util.Must(cluster.SetReplyHandler(handler))
	token.Init(com.SecretKey)

	// 初始化模块
	util.Must(playerMgr.Init(cfg.Ip, cfg.Port))

	// 初始化Actor
	safe.Go(func() { initApi(cfg) })
}

func Close() {
	playerMgr.Stop()
}

// 处理返回客户端的消息
func handler(head *pb.Head, body []byte) {
	mlog.Trace(head, "收到Nats数据包 body:%d", len(body))
	if len(head.Dst.ActorFunc) <= 0 {
		head.ActorName = "Player"
		head.FuncName = "SendToClient"
	} else {
		head.ActorName, head.FuncName = framework.ParseActorFunc(head.Dst.ActorFunc)
	}
	if head.Dst.ActorId <= 0 {
		head.ActorId = head.Dst.RouterId
	} else {
		head.ActorId = head.Dst.ActorId
	}
	if err := actor.Send(head, body); err != nil {
		mlog.Errorf("Actor消息转发失败: %v", err)
	}
}

func initApi(cfg *yaml.NodeConfig) error {
	api := http.NewServeMux()
	api.HandleFunc("/api/user/token", http_api.GenToken)
	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.HttpPort), api)
}
