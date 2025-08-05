package framework

import (
	"go-actor/common/pb"
	"go-actor/common/yaml"
	"go-actor/framework/actor"
	"go-actor/framework/cluster"
	"go-actor/framework/internal/request"
	"go-actor/framework/recycle"
	"go-actor/library/mlog"
	"go-actor/library/pprof"
	"go-actor/library/snowflake"
	"strings"
)

var (
	envType pb.EnvType
)

func GetEnvType() pb.EnvType {
	return envType
}

func Init(nn *pb.Node, srvCfg *yaml.NodeConfig, cfg *yaml.Config) error {
	if err := snowflake.Init(nn); err != nil {
		return err
	}
	// 初始化集群模块
	if err := cluster.Init(nn, srvCfg, cfg); err != nil {
		return err
	}
	initOther(cfg, nn)
	return nil
}

func InitDefault(nn *pb.Node, srvCfg *yaml.NodeConfig, cfg *yaml.Config) error {
	if err := Init(nn, srvCfg, cfg); err != nil {
		return err
	}
	// 初始化集群模块
	if err := cluster.SetBroadcastHandler(defaultHandler); err != nil {
		return err
	}
	if err := cluster.SetSendHandler(defaultHandler); err != nil {
		return err
	}
	if err := cluster.SetReplyHandler(defaultHandler); err != nil {
		return err
	}
	return nil
}

func initOther(cfg *yaml.Config, nn *pb.Node) {
	// 初始化全局变量
	switch strings.ToLower(cfg.Common.Env) {
	case "release":
		envType = pb.EnvType_EnvTypeRelease
	default:
		envType = pb.EnvType_EnvTypeDevelop
	}
	pprof.Init(cfg.Common, nn)
	recycle.Init()
}

func defaultHandler(head *pb.Head, buf []byte) {
	head.ActorName, head.FuncName = request.ParseActorFunc(head.Dst.ActorFunc)
	if head.Dst.ActorId <= 0 {
		head.ActorId = head.Dst.RouterId
	} else {
		head.ActorId = head.Dst.ActorId
	}
	if err := actor.Send(head, buf); err != nil {
		mlog.Error(head, "跨服务调用actor错误: %v", err)
	}
}
