package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"poker_server/framework"
	"poker_server/framework/token"
	"strconv"
)

type TokenRsp struct {
	Token string `json:"token"`
	Addr  string `json:"addr"`
}

func genToken(w http.ResponseWriter, r *http.Request) {
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
	// 返回token
	rsp := &TokenRsp{Token: tok, Addr: fmt.Sprintf("ws://%s/ws", framework.GetSelf().Addr)}
	buf, _ := json.Marshal(rsp)
	response(w, http.StatusOK, string(buf))
}
