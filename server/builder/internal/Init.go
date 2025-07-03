package internal

import (
	"go-actor/server/builder/internal/texas"
)

var (
	genMgr = texas.NewBuilderTexasGenerator()
)

func Init() error {
	if err := genMgr.Load(); err != nil {
		return err
	}
	return nil
}

func Close() {
	genMgr.Stop()
}
