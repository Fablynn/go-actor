package stat

import (
	"sync"
	"sync/atomic"
)

type Result struct {
	uid       uint64
	cmd       uint32
	flag      uint32
	startTime int64
	endTime   int64
}

type CmdStat struct {
	sync.WaitGroup
	total   int32
	cmd     uint32
	players map[uint64]*Result
}

func NewCmdStat(cmd uint32, uids ...uint64) *CmdStat {
	players := make(map[uint64]*Result)
	for _, uid := range uids {
		players[uid] = &Result{uid: uid, cmd: cmd}
	}
	return &CmdStat{players: players, cmd: cmd}
}

func (r *Result) Start(ms int64) {
	atomic.StoreInt64(&r.startTime, ms)
}

func (r *Result) Finish(ms int64, flag bool) {
	atomic.StoreInt64(&r.endTime, ms)
	if flag {
		atomic.AddUint32(&r.flag, 1)
	}
}

func (r *Result) IsFinish() bool {
	return atomic.LoadInt64(&r.endTime) > 0
}

func (r *CmdStat) Get(uid uint64) *Result {
	return r.players[uid]
}

func (r *CmdStat) Done() {
	atomic.AddInt32(&r.total, 1)
	r.WaitGroup.Done()
}
