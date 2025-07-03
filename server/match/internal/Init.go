package internal

import (
	"go-actor/server/match/internal/texas_room"
)

var (
	texasRoomMgr = texas_room.NewMatchTexasRoomMgr()
)

func Init() error {
	if err := texasRoomMgr.Load(); err != nil {
		return err
	}

	return nil
}

func Close() {
	texasRoomMgr.Close()
}
