package http

import (
	"go-actor/common/pb"
	"go-actor/framework"
	"net/http"
	"strconv"
)

func texasRoomList(w http.ResponseWriter, r *http.Request) {
	// 1. 只允许GET方法
	rsp := &pb.TexasRoomListRsp{}
	if r.Method != http.MethodGet {
		rsp.Head = &pb.RspHead{
			Code: int32(pb.ErrorCode_PARAM_INVALID),
			Msg:  "参数错误",
		}
		response(w, http.StatusMethodNotAllowed, rsp)
		return
	}
	// 2. 解析URL参数
	query := r.URL.Query()
	gameType, err := strconv.ParseInt(query.Get("GameType"), 10, 32)
	if err != nil {
		rsp.Head = &pb.RspHead{
			Code: int32(pb.ErrorCode_PARAM_INVALID),
			Msg:  "参数错误",
		}
		response(w, http.StatusBadRequest, rsp)
		return
	}
	coinType, err := strconv.ParseInt(query.Get("CoinType"), 10, 32)
	if err != nil {
		rsp.Head = &pb.RspHead{
			Code: int32(pb.ErrorCode_PARAM_INVALID),
			Msg:  "参数错误",
		}
		response(w, http.StatusBadRequest, rsp)
		return
	}

	req := &pb.TexasRoomListReq{
		GameType: int32(gameType),
		CoinType: int32(coinType),
	}
	dst := framework.NewMatchRouter(uint64(gameType)<<16|uint64(coinType), "MatchTexasRoom", "RoomListReq")
	if err := framework.Request(framework.NewHead(dst, 0, pb.RouterType_RouterTypeUid, 0), req, rsp); err != nil {
		rsp.Head = &pb.RspHead{
			Code: int32(pb.ErrorCode_PARAM_INVALID),
			Msg:  "请求失败",
		}
		response(w, http.StatusInternalServerError, rsp)
		return
	}
	response(w, http.StatusOK, rsp)
}
