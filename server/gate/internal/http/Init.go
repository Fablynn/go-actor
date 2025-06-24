package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"poker_server/common/yaml"

	"github.com/golang/protobuf/proto"
)

func Init(cfg *yaml.ServerConfig) error {
	api := http.NewServeMux()
	api.HandleFunc("/api/room/token", genToken)
	api.HandleFunc("/api/room/list", texasRoomList)
	api.HandleFunc("/api/game/buyin", buyInApi)
	api.HandleFunc("/api/room/sng/list", sngRoomList)
	api.HandleFunc("/api/room/rummy/list", rummyRoomList)
	api.HandleFunc("/api/game/reconnect", gameReconnect)
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
