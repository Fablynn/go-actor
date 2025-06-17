package signal

import (
	"os"
	"os/signal"
	"syscall"
)

func SignalNotify(ff func(), sigs ...os.Signal) {
	defaults := []os.Signal{syscall.SIGABRT, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}
	if len(sigs) > 0 {
		defaults = append(defaults, sigs...)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, defaults...)

	<-sig
	ff()
}
