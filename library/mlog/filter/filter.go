package filter

import (
	"fmt"
	"go-actor/common/pb"
)

var (
	filters = map[string]struct{}{
		"OnTick":       struct{}{},
		"HeartRequest": struct{}{},
	}
)

func IsFilter(head *pb.Head) bool {
	if head == nil {
		return true
	}
	if head.Src != nil {
		if _, ok := filters[head.Src.FuncName]; ok {
			return true
		}
	}
	if head.Dst != nil {
		if _, ok := filters[head.Dst.FuncName]; ok {
			return true
		}
	}
	_, ok := filters[head.FuncName]
	return ok
}

func Filter(head *pb.Head, format string) string {
	if head == nil {
		return format
	}
	if _, ok := filters[head.FuncName]; ok {
		return format
	}
	return fmt.Sprintf("SendType:%s, Src:%v, Dst:%v, Uid:%d, Seq:%d, Cmd:%d, Reply:%s", head.SendType, head.Src, head.Dst, head.Uid, head.Seq, head.Cmd, head.Reply, format)
}
