package pprof

import (
	"fmt"
	"go-actor/common/pb"
	"go-actor/common/yaml"
	"go-actor/library/mlog"
	"go-actor/library/safe"
	"net/http"
	"net/http/pprof"
	"os"
	rpprof "runtime/pprof"
	"strings"
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
func Init(cc *yaml.CommonConfig, nn *pb.Node) {
	if cc.Pprof {
		local.HandleFunc("/debug/pprof/", pprof.Index)
		local.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		local.HandleFunc("/debug/pprof/profile", pprof.Profile)
		local.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		local.HandleFunc("/debug/pprof/trace", pprof.Trace)

		safe.Go(func() {
			server := &http.Server{Addr: fmt.Sprintf(":%d", nn.Port+10000), Handler: local}
			if err := server.ListenAndServe(); err != nil {
				mlog.Errorf("pprof start failed, error:%v", err)
			}
		})
	}
}

func InitCpuProfile(nn *pb.Node) (*os.File, error) {
	cpuProfile, err := os.Create(fmt.Sprintf("%s%d.profile", strings.ToLower(nn.Name), nn.Id))
	if err != nil {
		return nil, err
	}
	if err := rpprof.StartCPUProfile(cpuProfile); err != nil {
		cpuProfile.Close()
		return nil, err
	}
	return cpuProfile, nil
}

func CloseCpuProfile(fb *os.File) {
	rpprof.StopCPUProfile()
	fb.Close()
}
