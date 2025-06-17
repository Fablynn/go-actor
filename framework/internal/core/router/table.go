package router

import (
	"fmt"
	"poker_server/common/pb"
	"poker_server/framework/domain"
	"poker_server/library/async"
	"poker_server/library/mlog"
	"sync"
	"time"
)

type Table struct {
	ttl   int64
	mutex sync.RWMutex
	exit  chan struct{}
	infos map[string]domain.IRouter
}

func New() *Table {
	return &Table{
		exit:  make(chan struct{}),
		infos: make(map[string]domain.IRouter),
	}
}

func (d *Table) Get(routerType pb.RouterType, id uint64) domain.IRouter {
	// 读取路由信息
	key := getKey(routerType, id)
	if info := d.get(key); info != nil {
		return info
	}
	// 创建路由信息
	val := &Router{updateTime: time.Now().Unix(), Router: &pb.Router{}}
	d.mutex.Lock()
	d.infos[key] = val
	d.mutex.Unlock()
	return val
}

func (d *Table) SetExpire(ttl int64) {
	d.ttl = ttl
	async.SafeGo(mlog.Fatalf, d.run)
}

func (r *Table) Close() {
	r.exit <- struct{}{}
}

func getKey(routerType pb.RouterType, id uint64) string {
	return fmt.Sprintf("%d:%d", routerType, id)
}

func (d *Table) get(key string) domain.IRouter {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	if router, ok := d.infos[key]; ok {
		return router
	}
	return nil
}

func (r *Table) run() {
	tt := time.NewTicker(time.Duration(r.ttl) * time.Second)
	defer tt.Stop()

	// 定时清理
	for {
		select {
		case <-tt.C:
			// 获取过期节点
			now := time.Now().Unix()
			if rets := r.getExpires(now); len(rets) > 0 {
				// 删除节点
				r.mutex.Lock()
				for _, val := range rets {
					delete(r.infos, val)
				}
				r.mutex.Unlock()
			}
		case <-r.exit:
			return
		}
	}
}

func (r *Table) getExpires(now int64) (rets []string) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for routeId, val := range r.infos {
		if val.IsExpire(now, r.ttl) {
			rets = append(rets, routeId)
		}
	}
	return
}
