package timer

import (
	"poker_server/common/pb"
	"poker_server/library/async"
	"poker_server/library/mlog"
	"poker_server/library/uerror"
	"sort"
	"sync/atomic"
	"time"
)

type Task struct {
	taskId *uint64
	task   func()
	ttl    int64
	expire int64
	times  int32
	next   *Task
}

type Wheel struct {
	mask    int64
	shift   int64
	cursor  int64
	buckets []*Task
}

type Timer struct {
	now       int64               // 当前时间
	startTime int64               // 启动时间
	head      *Wheel              // 头部时间轮
	tail      *Wheel              // 尾部时间轮
	wheels    []*Wheel            // 时间轮
	tasks     *async.Queue[*Task] // 待插入定时任务队列
	exit      chan struct{}       // 定时器退出通知
}

func NewTimer(tick int64) *Timer {
	now := time.Now().UnixMilli()
	ret := &Timer{
		now:       now,
		startTime: now,
		wheels: []*Wheel{
			{mask: (1 << 12) - 1, shift: tick, cursor: now, buckets: make([]*Task, 1<<12)},
			{mask: (1 << 5) - 1, shift: tick + 12, cursor: now, buckets: make([]*Task, 1<<5)},
			{mask: (1 << 5) - 1, shift: tick + 17, cursor: now, buckets: make([]*Task, 1<<5)},
			{mask: (1 << 5) - 1, shift: tick + 22, cursor: now, buckets: make([]*Task, 1<<5)},
			{mask: (1 << 5) - 1, shift: tick + 27, cursor: now, buckets: make([]*Task, 1<<5)},
		},
		tasks: async.NewQueue[*Task](),
		exit:  make(chan struct{}),
	}
	ret.head = ret.wheels[0]
	ret.tail = ret.wheels[len(ret.wheels)-1]
	async.SafeGo(mlog.Fatalf, ret.run)
	return ret
}

// 注册定时器
func (d *Timer) Register(taskId *uint64, f func(), ttl time.Duration, times int32) error {
	tt := int64(ttl / time.Millisecond)
	if tt>>d.head.shift <= 0 {
		return uerror.New(1, pb.ErrorCode_MIN_SIZE_LIMIT, "小于定时器最小时间间隔:%dms", 1<<d.head.shift)
	}
	if (tt >> d.tail.shift) > d.tail.mask {
		return uerror.New(1, pb.ErrorCode_MAX_SIZE_LIMIT, "定时器超出最大时间范围:%dms", d.tail.shift)
	}
	d.tasks.Push(&Task{
		taskId: taskId,
		task:   f,
		ttl:    tt,
		times:  times,
	})
	return nil
}

func (d *Timer) run() {
	tick := int64(1) << d.head.shift
	tt := time.NewTicker(time.Duration(tick) * time.Millisecond)
	defer func() {
		tt.Stop()
	}()

	for {
		select {
		case <-tt.C:
			now := atomic.AddInt64(&d.now, tick)
			d.update(now)
			d.flush(now)
		case <-d.exit:
			return
		}
	}
}

// 刷入新定时器
func (d *Timer) flush(now int64) {
	news := []*Task{}
	for tt := d.tasks.Pop(); tt != nil; tt = d.tasks.Pop() {
		tt.expire = now + tt.ttl
		news = append(news, tt)
	}

	sort.Slice(news, func(i, j int) bool {
		return news[i].expire < news[j].expire
	})

	pos := 0
	lnews := len(news)
	for _, w := range d.wheels {
		for ; pos < lnews && w.IsMatch(news[pos]); pos++ {
			w.Insert(news[pos])
		}
		if lnews <= pos {
			break
		}
	}
}

func (d *Timer) update(now int64) {
	news := []*Task{}
	for _, w := range d.wheels {
		for tts := w.Get(now); tts != nil; {
			tt := tts
			tts = tts.next
			tt.next = nil
			if !d.head.IsExpire(tt) {
				news = append(news, tt)
			} else {
				if newTask := tt.handle(now); newTask != nil {
					news = append(news, newTask)
				}
			}
		}
		if !w.IsCarry() {
			break
		}
	}

	sort.Slice(news, func(i, j int) bool {
		return news[i].expire < news[j].expire
	})

	pos := 0
	lnews := len(news)
	for _, w := range d.wheels {
		for ; pos < lnews && w.IsMatch(news[pos]); pos++ {
			w.Insert(news[pos])
		}
		if lnews <= pos {
			break
		}
	}
}

// 是否进位
func (w *Wheel) IsCarry() bool {
	return (w.cursor>>w.shift)&w.mask <= 0
}

// 是否过期
func (w *Wheel) IsExpire(tt *Task) bool {
	return tt.expire <= w.cursor || (tt.expire>>w.shift) <= (w.cursor>>w.shift)
}

// 是否匹配
func (w *Wheel) IsMatch(tt *Task) bool {
	return (tt.expire>>w.shift)-(w.cursor>>w.shift) <= w.mask
}

// 插入数据
func (w *Wheel) Insert(tt *Task) {
	pos := (tt.expire >> w.shift) & w.mask
	tt.next = w.buckets[pos]
	w.buckets[pos] = tt
}

// 获取过期定时任务
func (w *Wheel) Get(now int64) *Task {
	pos := (now >> w.shift) & w.mask
	ret := w.buckets[pos]
	w.buckets[pos] = nil
	w.cursor = now
	return ret
}

func (tt *Task) handle(now int64) *Task {
	if *tt.taskId <= 0 || tt.times == 0 {
		return nil
	}
	tt.task()
	if tt.times > 0 {
		tt.times--
	}
	if tt.times == 0 {
		return nil
	}
	tt.expire = now + tt.ttl
	return tt
}

var (
	tt = NewTimer(4)
)

func Register(taskId *uint64, f func(), ttl time.Duration, times int32) error {
	return tt.Register(taskId, f, ttl, times)
}
