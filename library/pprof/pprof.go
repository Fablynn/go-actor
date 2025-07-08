package pprof

import (
	"fmt"
	"go-actor/library/mlog"
	"go-actor/library/safe"
	"net/http"
	"net/http/pprof"
)

var (
	local = http.NewServeMux()
)

func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	local.HandleFunc(pattern, handler)
}

func Handle(pattern string, handler http.Handler) {
	local.Handle(pattern, handler)
}

// 本地服务端口
func Init(ip string, port int) {
	safe.Go(func() {
		server := &http.Server{Addr: fmt.Sprintf("%s:%d", ip, port), Handler: local}
		if err := server.ListenAndServe(); err != nil {
			mlog.Errorf("pprof start failed, error:%v", err)
		}
	})
}

// 默认开启pprof工具
func init() {
	local.HandleFunc("/debug/pprof/", pprof.Index)
	local.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	local.HandleFunc("/debug/pprof/profile", pprof.Profile)
	local.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	local.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
