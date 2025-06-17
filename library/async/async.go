package async

import (
	"sync"
	"sync/atomic"
)

type Async struct {
	sync.WaitGroup
	id     uint64         // 唯一id
	status int32          // actor运行状态
	queue  *Queue[func()] // 任务队列
	notify chan struct{}  // 消耗通知
	exit   chan struct{}  // 退出
}

func NewAsync() *Async {
	return &Async{
		status: 0,
		queue:  NewQueue[func()](),
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

func (d *Async) Push(f func()) {
	if atomic.LoadInt32(&d.status) > 0 {
		d.queue.Push(f)
		// 避免阻塞
		select {
		case d.notify <- struct{}{}:
		default:
		}
	}
}

func (d *Async) Stop() {
	if atomic.LoadInt32(&d.status) <= 0 {
		return
	}
	atomic.StoreInt32(&d.status, 0)
	d.id = 0
	// 等待停止
	d.exit <- struct{}{}
	d.Wait()
}

func (d *Async) Start() {
	if atomic.LoadInt32(&d.status) > 0 {
		return
	}
	atomic.AddInt32(&d.status, 1)
	SafeGo(nil, d.run)
}

func (d *Async) run() {
	defer func() {
		for f := d.queue.Pop(); f != nil; f = d.queue.Pop() {
			SafeRecover(catch, f)
		}
		d.Done()
	}()
	d.Add(1)

	for {
		select {
		case <-d.notify:
			for f := d.queue.Pop(); f != nil; f = d.queue.Pop() {
				SafeRecover(catch, f)
			}
		case <-d.exit:
			return
		}
	}
}
