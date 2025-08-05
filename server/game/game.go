package main

import (
	"flag"
	"fmt"
	"go-actor/common/config"
	"go-actor/common/pb"
	"go-actor/common/redis"
	"go-actor/common/yaml"
	"go-actor/framework"
	"go-actor/framework/cluster"
	"go-actor/framework/recycle"
	"go-actor/library/mlog"
	"go-actor/library/signal"
	"go-actor/library/util"
	"go-actor/message"
	"go-actor/server/game/internal"
)

func main() {
	var cfg string
	var nodeId int
	flag.StringVar(&cfg, "config", "config.yaml", "游戏配置文件")
	flag.IntVar(&nodeId, "id", 1, "服务ID")
	flag.Parse()

	// 加载游戏配置
	yamlcfg, node, err := yaml.LoadConfig(cfg, pb.NodeType_NodeTypeGame, int32(nodeId))
	if err != nil {
		panic(fmt.Sprintf("游戏配置加载失败: %v", err))
	}
	nodeCfg := yamlcfg.Game[node.Id]

	// 初始化日志库
	mlog.Init(node.Name, node.Id, nodeCfg.LogLevel, nodeCfg.LogPath)

	// 初始化游戏配置
	mlog.Infof("初始化游戏配置")
	util.Must(config.Init(yamlcfg.Etcd, yamlcfg.Data))

	// 初始化redis
	mlog.Infof("初始化redis配置")
	util.Must(redis.Init(yamlcfg.Redis))

	// 初始化框架
	mlog.Infof("启动框架服务")
	util.Must(framework.InitDefault(node, nodeCfg, yamlcfg))

	// 功能模块初始化 todo
	mlog.Infof("初始化功能模块")
	message.Init()

	// 服务退出
	signal.SignalNotify(func() {
		recycle.Close()
		internal.Close()
		cluster.Close()
		mlog.Close()
	})
}
