package http

import (
	"encoding/json"
	"fmt"
	"go-actor/common/yaml"
	"net/http"

	"github.com/golang/protobuf/proto"
)

func Init(cfg *yaml.ServerConfig) error {
	api := http.NewServeMux()
	api.HandleFunc("/api/room/token", genToken)
	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.HttpPort), api)
}

func response(w http.ResponseWriter, code int, rsp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	switch vv := rsp.(type) {
	case proto.Message:
		buf, _ := json.Marshal(rsp)
		w.Write(buf)
	case string:
		w.Write([]byte(vv))
	}
}
