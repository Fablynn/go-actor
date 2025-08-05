package http_api

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"go-actor/common/token"
	"go-actor/framework/cluster"
	"net/http"
	"strconv"
)

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

type TokenRsp struct {
	Token string `json:"token"`
	Addr  string `json:"addr"`
}

func GenToken(w http.ResponseWriter, r *http.Request) {
	// 1. 只允许POST方法
	if r.Method != http.MethodGet {
		response(w, http.StatusMethodNotAllowed, "只支持get方法")
		return
	}
	// 2. 解析URL参数
	query := r.URL.Query()
	uid, err := strconv.ParseUint(query.Get("Uid"), 10, 64)
	if err != nil {
		response(w, http.StatusBadRequest, err.Error())
		return
	}
	roomId, err := strconv.ParseUint(query.Get("RoomId"), 10, 64)
	if err != nil {
		response(w, http.StatusBadRequest, err.Error())
		return
	}
	// 生成token
	tok, err := token.GenToken(&token.Token{RoomId: roomId, Uid: uid})
	if err != nil {
		response(w, http.StatusBadRequest, err.Error())
		return
	}

	addr := cluster.GetSelf().Ip
	// 返回token
	rsp := &TokenRsp{Token: tok, Addr: addr}
	buf, _ := json.Marshal(rsp)
	response(w, http.StatusOK, string(buf))
}
