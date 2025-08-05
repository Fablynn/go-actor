package actor

import (
	"go-actor/common/pb"
	"go-actor/framework/define"
	"go-actor/library/timer"
	"go-actor/library/uerror"
	"reflect"
	"strings"
	"sync/atomic"
	"time"
)

var (
	actors = make(map[string]define.IActor)
	t      = timer.NewTimer(4)
	uuid   = uint64(0)
)

func GenActorId() uint64 {
	return atomic.AddUint64(&uuid, 1)
}

func Register(ac define.IActor) {
	actors[ac.GetActorName()] = ac
}

func SendMsg(head *pb.Head, args ...interface{}) error {
	if act, ok := actors[head.ActorName]; ok {
		return act.SendMsg(head, args...)
	}
	return uerror.New(pb.ErrorCode_ACTOR_NOT_SUPPORTED, "Actor(%s)不存在", head.ActorName)
}

func Send(head *pb.Head, body []byte) error {
	if act, ok := actors[head.ActorName]; ok {
		return act.Send(head, body)
	}
	return uerror.New(pb.ErrorCode_ACTOR_NOT_SUPPORTED, "Actor(%s)不存在", head.ActorName)
}

func RegisterTimer(id *uint64, f func(), ttl time.Duration, times int32) error {
	return t.Register(id, f, ttl, times)
}

func parseName(rr reflect.Type) string {
	name := rr.String()
	if index := strings.Index(name, "."); index > -1 {
		name = name[index+1:]
	}
	return name
}
