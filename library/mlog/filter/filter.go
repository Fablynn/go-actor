package filter

import (
	"fmt"
	"poker_server/common/pb"
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
	if _, ok := filters[head.FuncName]; ok {
		return true
	}
	return false
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
