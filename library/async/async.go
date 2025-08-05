package async

import (
	"go-actor/library/queue"
	"go-actor/library/safe"
	"sync"
	"sync/atomic"
)

type Async struct {
	sync.WaitGroup
	id     uint64               // 唯一id
	status int32                // actor运行状态
	tasks  *queue.Queue[func()] // 任务队列
	notify chan struct{}        // 消耗通知
	exit   chan struct{}        // 退出
}

func NewAsync() *Async {
	return &Async{
		tasks:  queue.NewQueue[func()](),
		notify: make(chan struct{}, 1),
		exit:   make(chan struct{}, 1),
	}
}

func (a *Async) GetIdPointer() *uint64 {
	return &a.id
}

func (a *Async) GetId() uint64 {
	return a.id
}

func (a *Async) SetId(id uint64) {
	a.id = id
}

func (d *Async) Stop() {
	if atomic.LoadInt32(&d.status) <= 0 {
		return
	}
	atomic.StoreInt32(&d.status, 0)
	close(d.exit)
	d.Wait()
	atomic.StoreUint64(&d.id, 0)
}

func (d *Async) Start() {
	if atomic.LoadInt32(&d.status) > 0 {
		return
	}
	d.Add(1)
	atomic.AddInt32(&d.status, 1)
	go d.run()
}

func (d *Async) Push(f func()) {
	if atomic.LoadInt32(&d.status) > 0 {
		d.tasks.Push(f)
		select {
		case d.notify <- struct{}{}:
		default:
		}
	}
}

func (d *Async) run() {
	defer func() {
		for f := d.tasks.Pop(); f != nil; f = d.tasks.Pop() {
			safe.Recover(f)
		}
		d.Done()
	}()
	for {
		select {
		case <-d.notify:
			for f := d.tasks.Pop(); f != nil; f = d.tasks.Pop() {
				safe.Recover(f)
			}
		case <-d.exit:
			return
		}
	}
}
