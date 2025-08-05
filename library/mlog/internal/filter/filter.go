package filter

import (
	"fmt"
	"go-actor/common/pb"
	"strings"
)

var (
	filters = map[string]struct{}{
		"OnTick":       struct{}{},
		"HeartRequest": struct{}{},
	}
)

func IsFilter(head *pb.Head) bool {
	if head == nil {
		return false
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
	if head.Src == nil && head.Dst == nil {
		return fmt.Sprintf("Uid(%d), Cmd(%d), Seq(%d), SendType(%s), Reply(%s)\t%s", head.Uid, head.Cmd, head.Seq, head.SendType, head.Reply, format)
	}
	return fmt.Sprintf("[%s -> %s] Uid(%d), Cmd(%d), Seq(%d), SendType(%s), Reply(%s)\t%s", ToString(head.Src), ToString(head.Dst), head.Uid, head.Cmd, head.Seq, head.SendType, head.Reply, format)
}

func ToString(nn *pb.NodeRouter) string {
	if nn == nil {
		return ""
	}
	nodeType := strings.TrimPrefix(nn.NodeType.String(), "NodeType")
	return fmt.Sprintf("%s%d.%s(%d,%d,%d)", nodeType, nn.NodeId, nn.ActorFunc, nn.RouterType, nn.RouterId, nn.ActorId)
}
