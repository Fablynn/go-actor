package safe

import (
	"go-actor/library/mlog"
	"runtime/debug"
)

func Recover(f func()) {
	defer func() {
		if err := recover(); err != nil {
			mlog.Fatalf("%v stack: %v", err, string(debug.Stack()))
		}
	}()
	f()
}

func Go(f func()) {
	go func() {
		Recover(f)
	}()
}
