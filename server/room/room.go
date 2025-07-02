package main

import (
	"flag"
	"fmt"
	"go-actor/common/config"
	"go-actor/common/dao"
	"go-actor/common/pb"
	"go-actor/common/yaml"
	"go-actor/framework"
	"go-actor/library/async"
	"go-actor/library/mlog"
	"go-actor/library/signal"
	"go-actor/server/room/internal/manager"
)

func main() {
	var cfg string
	var nodeId int
	flag.StringVar(&cfg, "config", "config.yaml", "游戏配置文件")
	flag.IntVar(&nodeId, "id", 1, "服务ID")
	flag.Parse()

	// 加载游戏配置
	yamlcfg, node, err := yaml.LoadConfig(cfg, pb.NodeType_NodeTypeRoom, int32(nodeId))
	if err != nil {
		panic(fmt.Sprintf("游戏配置加载失败: %v", err))
	}
	nodeCfg := yamlcfg.Room[node.Id]

	// 初始化日志库
	if err := mlog.Init(yamlcfg.Common.Env, nodeCfg.LogLevel, nodeCfg.LogFile); err != nil {
		panic(fmt.Sprintf("日志库初始化失败: %v", err))
	}
	async.Init(mlog.Errorf)

	// 初始化游戏配置
	mlog.Infof("初始化游戏配置")
	if err := config.Init(yamlcfg.Etcd, yamlcfg.Common); err != nil {
		panic(err)
	}

	// 初始化redis
	mlog.Infof("初始化redis配置")
	if err := dao.InitRedis(yamlcfg.Redis); err != nil {
		panic(fmt.Sprintf("redis初始化失败: %v", err))
	}

	// 初始化框架
	mlog.Infof("启动框架服务: %v", node)
	if err := framework.InitDefault(node, nodeCfg, yamlcfg); err != nil {
		panic(fmt.Sprintf("框架初始化失败: %v", err))
	}

	// 功能模块初始化 todo
	if err := manager.Init(); err != nil {
		panic(fmt.Sprintf("功能模块初始化失败: %v", err))
	}

	// 服务退出
	signal.SignalNotify(func() {
		manager.Close()
		framework.Close()
		mlog.Close()
	})
}
