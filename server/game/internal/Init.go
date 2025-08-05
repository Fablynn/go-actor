package internal

import (
	"go-actor/server/game/internal/player"
	"go-actor/server/game/internal/prop"
)

var (
	propMgr   = prop.NewPropMgr()
	playerMgr = player.NewPlayerMgr()
)

func Close() {
	propMgr.Close()
	playerMgr.Close()
}
