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
	return fmt.Sprintf("%s->%s, SendType:%s, Uid:%d, Seq:%d, Cmd:%d, Reply:%s | %s", ToString(head.Src), ToString(head.Dst), head.SendType, head.Uid, head.Seq, head.Cmd, head.Reply, format)
}

func ToString(nn *pb.NodeRouter) string {
	if nn == nil {
		return ""
	}
	nodeType := strings.TrimPrefix(nn.NodeType.String(), "NodeType")
	return fmt.Sprintf("%s%d(%d).%s.%s(%d)", nodeType, nn.NodeId, nn.RouterType, nn.ActorName, nn.FuncName, nn.ActorId)
}
