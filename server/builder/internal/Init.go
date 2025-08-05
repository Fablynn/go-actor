package internal

import (
	"go-actor/library/util"
	"go-actor/server/builder/internal/gen"
)

var (
	genMgr = gen.NewGenerator()
)

func Init() {
	util.Must(genMgr.Init())
}

func Close() {
	genMgr.Stop()
}
