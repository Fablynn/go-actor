package main

import (
	"flag"
	"fmt"
	"net/http"
	"poker_server/common/config"
	"poker_server/common/pb"
	"poker_server/common/yaml"
	"poker_server/framework"
	"poker_server/framework/token"
	"poker_server/library/async"
	"poker_server/library/mlog"
	"poker_server/library/signal"
	"poker_server/server/client/internal/manager"

	"github.com/spf13/cast"
)

var (
	playerMgr *manager.ClientPlayerMgr
)

func main() {
	var filename string
	var id, port int
	var begin, end int64
	flag.StringVar(&filename, "config", "local.yaml", "游戏配置")
	flag.IntVar(&id, "id", 1, " 节点id")
	flag.IntVar(&port, "port", 22345, " 节点端口")
	flag.Int64Var(&begin, "begin", 144, "起始uid")
	flag.Int64Var(&end, "end", 145, "终止uid")
	flag.Parse()

	// 加载游戏配置
	yamlcfg, node, err := yaml.LoadConfig(filename, pb.NodeType_NodeTypeClient, int32(id))
	if err != nil {
		panic(fmt.Sprintf("游戏配置加载失败: %v", err))
	}
	nodeCfg := yamlcfg.Client[node.Id]

	// 初始化日志库
	if err := mlog.Init(yamlcfg.Common.Env, nodeCfg.LogLevel, nodeCfg.LogFile); err != nil {
		panic(fmt.Sprintf("日志库初始化失败: %v", err))
	}
	async.Init(mlog.Errorf)
	token.Init(yamlcfg.Common.SecretKey)

	// 配置初始化
	if err := config.Init(yamlcfg.Etcd, yamlcfg.Common); err != nil {
		panic(fmt.Sprintf("配置初始化失败: %v", err))
	}

	// 初始化PlayerMgr
	playerMgr = manager.NewClientPlayerMgr(node, nodeCfg)
	playerMgr.Login(uint64(begin), uint64(end))

	// 请求 http 服务，接受请求
	http.HandleFunc("/api", handle)
	async.SafeGo(mlog.Errorf, func() {
		http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	})

	signal.SignalNotify(func() {
		framework.Close()
		mlog.Close()
	})
}

func handle(w http.ResponseWriter, r *http.Request) {
	// 1. 只允许POST方法
	if r.Method != http.MethodGet {
		http.Error(w, "只支持post方法", http.StatusMethodNotAllowed)
		return
	}
	// 2. 解析URL参数
	query := r.URL.Query()
	cmd := query.Get("cmd")
	if len(cmd) <= 0 {
		http.Error(w, "cmd不能为空", http.StatusBadRequest)
		return
	}
	var routeId uint64
	rr := query.Get("route_id")
	if len(rr) <= 0 {
		routeId = 0
	} else {
		routeId = cast.ToUint64(rr)
	}
	jsonVal := query.Get("json")
	uid := uint64(0)
	if uidStr := query.Get("uid"); len(uidStr) > 0 {
		uid = cast.ToUint64(uidStr)
	}

	playerMgr.SendCmd(cast.ToUint32(cmd), uid, routeId, jsonVal)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}
