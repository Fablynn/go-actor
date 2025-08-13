package recycle

import (
	"go-actor/framework/define"
	"go-actor/library/queue"
	"go-actor/library/safe"
	"sync"
)

var (
	gc *Recycle
)

type IDestroy interface {
	Close()
}

type WrapDestroy struct {
	define.IActor
}

func (w *WrapDestroy) Close() {
	w.Stop()
}

type Recycle struct {
	sync.WaitGroup
	tasks  *queue.Queue[IDestroy]
	notify chan struct{} // 消耗通知
	exit   chan struct{} // 退出
}

func Init() {
	gc = &Recycle{
		tasks:  queue.NewQueue[IDestroy](),
		notify: make(chan struct{}, 1),
		exit:   make(chan struct{}, 1),
	}
	gc.Add(1)
	safe.Go(run)
}

func Close() {
	close(gc.exit)
	gc.Wait()
}

func DestroyActor(acts ...define.IActor) {
	for _, act := range acts {
		gc.tasks.Push(&WrapDestroy{IActor: act})
	}

	select {
	case gc.notify <- struct{}{}:
	default:
	}
}

func Destroy(fs ...IDestroy) {
	for _, f := range fs {
		gc.tasks.Push(f)
	}

	select {
	case gc.notify <- struct{}{}:
	default:
	}
}

func run() {
	defer func() {
		for f := gc.tasks.Pop(); f != nil; f = gc.tasks.Pop() {
			safe.Recover(f.Close)
		}
		gc.Done()
	}()
	for {
		select {
		case <-gc.notify:
			for f := gc.tasks.Pop(); f != nil; f = gc.tasks.Pop() {
				safe.Recover(f.Close)
			}
		case <-gc.exit:
			return
		}
	}
}
