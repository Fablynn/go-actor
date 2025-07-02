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
	src, dst := head.Src, head.Dst
	if src == nil || dst == nil {
		return fmt.Sprintf("Actor(%s.%s) %s", head.ActorName, head.FuncName, format)
	}
	return fmt.Sprintf("[%s.%s.%s(%d) -> %s.%s.%s(%d)] %s", src.NodeType.String(), src.ActorName, src.FuncName, src.NodeId,
		dst.NodeType.String(), dst.ActorName, dst.FuncName, dst.NodeId, format)
}
