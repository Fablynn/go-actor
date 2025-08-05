package factory

import (
	"go-actor/common/pb"
	"go-actor/server/game/internal/player/domain"
	"go-actor/server/game/internal/player/playerfun"
)

var (
	LoadList = []pb.PlayerDataType{
		pb.PlayerDataType_PLAYER_DATA_INFO,
		pb.PlayerDataType_PLAYER_DATA_ROOM,
		pb.PlayerDataType_PLAYER_DATA_BAG,
	}
	FUNCS = make(map[pb.PlayerDataType]func(*playerfun.PlayerFun) domain.IPlayerFun)
)

func init() {
	FUNCS[pb.PlayerDataType_PLAYER_DATA_INFO] = playerfun.NewPlayerInfoFun
	FUNCS[pb.PlayerDataType_PLAYER_DATA_ROOM] = playerfun.NewPlayerRoomFun
	FUNCS[pb.PlayerDataType_PLAYER_DATA_BAG] = playerfun.NewPlayerBagFun
}
