package router

import (
	"go-actor/common/pb"
	"go-actor/framework/define"
	"sync/atomic"
	"time"
)

type Router struct {
	*pb.Router
	updateTime int64
}

func (d *Router) GetUpdateTime() int64 {
	return atomic.LoadInt64(&d.updateTime)
}

func (d *Router) SetUpdateTime() {
	atomic.StoreInt64(&d.updateTime, time.Now().Unix())
}

func (d *Router) GetData() *pb.Router {
	d.SetUpdateTime()
	return &pb.Router{
		Gate:    atomic.LoadInt32(&d.Gate),
		Game:    atomic.LoadInt32(&d.Game),
		Room:    atomic.LoadInt32(&d.Room),
		Match:   atomic.LoadInt32(&d.Match),
		Db:      atomic.LoadInt32(&d.Db),
		Builder: atomic.LoadInt32(&d.Builder),
		Gm:      atomic.LoadInt32(&d.Gm),
	}
}

func (d *Router) SetData(info *pb.Router) define.IRouter {
	if info != nil {
		d.SetUpdateTime()
		d.Set(pb.NodeType_NodeTypeGate, info.Gate)
		d.Set(pb.NodeType_NodeTypeRoom, info.Room)
		d.Set(pb.NodeType_NodeTypeMatch, info.Match)
		d.Set(pb.NodeType_NodeTypeDb, info.Db)
		d.Set(pb.NodeType_NodeTypeBuilder, info.Builder)
		d.Set(pb.NodeType_NodeTypeGm, info.Gm)
		d.Set(pb.NodeType_NodeTypeGame, info.Game)
	}
	return d
}

func (d *Router) Get(nodeType pb.NodeType) int32 {
	d.SetUpdateTime()
	switch nodeType {
	case pb.NodeType_NodeTypeRoom:
		return atomic.LoadInt32(&d.Room)
	case pb.NodeType_NodeTypeMatch:
		return atomic.LoadInt32(&d.Match)
	case pb.NodeType_NodeTypeDb:
		return atomic.LoadInt32(&d.Db)
	case pb.NodeType_NodeTypeBuilder:
		return atomic.LoadInt32(&d.Builder)
	case pb.NodeType_NodeTypeGm:
		return atomic.LoadInt32(&d.Gm)
	case pb.NodeType_NodeTypeGame:
		return atomic.LoadInt32(&d.Game)
	default:
		return atomic.LoadInt32(&d.Gate)
	}
}

func (d *Router) Set(nodeType pb.NodeType, nodeId int32) define.IRouter {
	if nodeId > 0 {
		d.SetUpdateTime()
		switch nodeType {
		case pb.NodeType_NodeTypeGate:
			atomic.StoreInt32(&d.Gate, nodeId)
		case pb.NodeType_NodeTypeRoom:
			atomic.StoreInt32(&d.Room, nodeId)
		case pb.NodeType_NodeTypeMatch:
			atomic.StoreInt32(&d.Match, nodeId)
		case pb.NodeType_NodeTypeDb:
			atomic.StoreInt32(&d.Db, nodeId)
		case pb.NodeType_NodeTypeBuilder:
			atomic.StoreInt32(&d.Builder, nodeId)
		case pb.NodeType_NodeTypeGm:
			atomic.StoreInt32(&d.Gm, nodeId)
		case pb.NodeType_NodeTypeGame:
			atomic.StoreInt32(&d.Game, nodeId)
		}
	}
	return d
}
