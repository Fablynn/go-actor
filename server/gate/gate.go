package main

import (
	"flag"
	"fmt"
	"poker_server/common/config"
	"poker_server/common/dao"
	"poker_server/common/pb"
	"poker_server/common/yaml"
	"poker_server/framework"
	"poker_server/library/async"
	"poker_server/library/mlog"
	"poker_server/library/signal"
	"poker_server/server/gate/internal/manager"
)

func main() {
	var cfg string
	var nodeId int
	flag.StringVar(&cfg, "config", "config.yaml", "游戏配置文件")
	flag.IntVar(&nodeId, "id", 1, "服务ID")
	flag.Parse()

	// 加载游戏配置
	yamlCfg, node, err := yaml.LoadConfig(cfg, pb.NodeType_NodeTypeGate, int32(nodeId))
	if err != nil {
		panic(fmt.Sprintf("游戏配置加载失败: %v", err))
	}
	nodeCfg := yamlCfg.Gate[node.Id]

	// 初始化日志库
	if err := mlog.Init(yamlCfg.Common.Env, nodeCfg.LogLevel, nodeCfg.LogFile); err != nil {
		panic(fmt.Sprintf("日志库初始化失败: %v", err))
	}
	async.Init(mlog.Errorf)

	// 初始化redis
	mlog.Infof("初始化redis连接池")
	if err := dao.InitRedis(yamlCfg.Redis); err != nil {
		panic(fmt.Sprintf("redis初始化失败: %v", err))
	}

	// 初始化配置
	mlog.Infof("初始化游戏配置")
	if err := config.Init(yamlCfg.Etcd, yamlCfg.Common); err != nil {
		panic(fmt.Sprintf("游戏配置初始化失败: %v", err))
	}

	// 初始化框架核心
	mlog.Infof("启动框架服务: %v", node)
	if err := framework.Init(node, nodeCfg, yamlCfg); err != nil {
		panic(fmt.Sprintf("框架核心初始化失败: %v", err))
	}
	mlog.Infof("框架核心初始化成功s")

	// 初始化模块+websocket服务
	mlog.Infof("初始化模块")
	if err := manager.Init(nodeCfg, yamlCfg.Common); err != nil {
		panic(fmt.Sprintf("模块初始化失败: %v", err))
	}

	// 服务退出
	signal.SignalNotify(func() {
		manager.Close()
		framework.Close()
		mlog.Close()
	})
}
