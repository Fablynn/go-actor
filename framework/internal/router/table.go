package router

import (
	"go-actor/common/pb"
	"go-actor/framework/define"
	"go-actor/library/mlog"
	"go-actor/library/safe"
	"sync"
	"time"
)

type Table struct {
	mutex sync.RWMutex
	data  map[uint64]*Router
	exit  chan struct{}
	ttl   int64
}

func New(ttl int64) *Table {
	ret := &Table{
		exit: make(chan struct{}),
		data: make(map[uint64]*Router),
		ttl:  ttl,
	}
	safe.Go(ret.run)
	return ret
}

func getkey(routerType pb.RouterType, id uint64) uint64 {
	return uint64(routerType)<<56 | id&0xFFFFFFFFFFFFFF
}

func (d *Table) Add(routerType pb.RouterType, id uint64, nn *pb.Node, rr *pb.Router) {
	tt := d.GetOrNew(routerType, id, nn)
	tt.SetData(rr)
}

func (d *Table) Get(routerType pb.RouterType, id uint64) define.IRouter {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	if val, ok := d.data[getkey(routerType, id)]; ok {
		return val
	}
	return nil
}

func (d *Table) GetOrNew(routerType pb.RouterType, id uint64, nn *pb.Node) define.IRouter {
	if rr := d.Get(routerType, id); rr != nil {
		return rr
	}

	// 创建路由信息
	val := &Router{updateTime: time.Now().Unix(), Router: &pb.Router{}}
	val.Set(nn.Type, nn.Id)

	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.data[getkey(routerType, id)] = val
	return val
}

func (r *Table) Close() {
	close(r.exit)
}

func (t *Table) Walk(f func(uint64, *Router) bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	for id, rr := range t.data {
		if !f(id, rr) {
			return
		}
	}
}

func (t *Table) Remove(ids ...uint64) {
	if len(ids) <= 0 {
		return
	}
	mlog.Infof("删除路由：%v", ids)
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for _, id := range ids {
		delete(t.data, id)
	}
}

func (t *Table) run() {
	tt := time.NewTicker(15 * time.Second)
	defer tt.Stop()
	for {
		select {
		case <-tt.C:
			now := time.Now().Unix()
			dels := []uint64{}
			t.Walk(func(id uint64, rr *Router) bool {
				if t.ttl < now-rr.GetUpdateTime() {
					dels = append(dels, id)
				}
				return true
			})
			t.Remove(dels...)
		case <-t.exit:
			return
		}
	}
}
