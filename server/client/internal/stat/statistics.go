package stat

import (
	"go-actor/common/pb"
	"go-actor/framework/actor"
	"go-actor/library/mlog"
	"reflect"
	"time"
)

type Statistics struct {
	actor.Actor
}

func NewStatistics() *Statistics {
	ret := &Statistics{}
	ret.Actor.Register(ret)
	ret.Actor.ParseFunc(reflect.TypeOf(ret))
	ret.Actor.Start()
	actor.Register(ret)
	return ret
}

func (d *Statistics) Analysis(st *CmdStat) {
	time.Sleep(2 * time.Second)
	st.Wait()
	var success, fail, overs, sum, min, max int64
	for _, ret := range st.players {
		if ret.flag > 0 {
			success++
		} else {
			fail++
		}
		diff := ret.endTime - ret.startTime
		if min == 0 {
			min = diff
		}
		if diff > 200 {
			overs++
		}
		if min > diff {
			min = diff
		}
		if max < diff {
			max = diff
		}
		sum += diff
	}
	avg := sum / int64(st.total)
	mlog.Infof("%s ---> 总请求:%d, 超时请求(超过200ms的请求):%d, 成功:%d, 失败:%d, avg:%dms, min:%dms, max:%dms", pb.CMD(st.cmd), st.total, overs, success, fail, avg, min, max)
}
