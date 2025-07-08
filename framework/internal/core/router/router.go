package router

import (
	"go-actor/common/pb"
	"sync/atomic"
	"time"
)

type Router struct {
	*pb.Router
	updateTime int64
}

func (d *Router) GetData() *pb.Router {
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

func (d *Router) IsExpire(now, ttl int64) bool {
	return now >= atomic.LoadInt64(&d.updateTime)+ttl
}

func (d *Router) Get(nodeType pb.NodeType) int32 {
	switch nodeType {
	case pb.NodeType_NodeTypeGate:
		return atomic.LoadInt32(&d.Gate)
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
	}
	return d.Gate
}

func (d *Router) Set(nodeType pb.NodeType, nodeId int32) {
	if nodeId > 0 {
		switch nodeType {
		case pb.NodeType_NodeTypeGate:
			atomic.StoreInt32(&d.Gate, nodeId)
			atomic.StoreInt64(&d.updateTime, time.Now().Unix())
		case pb.NodeType_NodeTypeRoom:
			atomic.StoreInt32(&d.Room, nodeId)
			atomic.StoreInt64(&d.updateTime, time.Now().Unix())
		case pb.NodeType_NodeTypeMatch:
			atomic.StoreInt32(&d.Match, nodeId)
			atomic.StoreInt64(&d.updateTime, time.Now().Unix())
		case pb.NodeType_NodeTypeDb:
			atomic.StoreInt32(&d.Db, nodeId)
			atomic.StoreInt64(&d.updateTime, time.Now().Unix())
		case pb.NodeType_NodeTypeBuilder:
			atomic.StoreInt32(&d.Builder, nodeId)
			atomic.StoreInt64(&d.updateTime, time.Now().Unix())
		case pb.NodeType_NodeTypeGm:
			atomic.StoreInt32(&d.Gm, nodeId)
			atomic.StoreInt64(&d.updateTime, time.Now().Unix())
		case pb.NodeType_NodeTypeGame:
			atomic.StoreInt32(&d.Game, nodeId)
			atomic.StoreInt64(&d.updateTime, time.Now().Unix())
		}
	}
}

func (d *Router) SetData(info *pb.Router) {
	if info == nil {
		return
	}
	if info.Gate > 0 {
		d.Set(pb.NodeType_NodeTypeGate, info.Gate)
	}
	if info.Room > 0 {
		d.Set(pb.NodeType_NodeTypeRoom, info.Room)
	}
	if info.Match > 0 {
		d.Set(pb.NodeType_NodeTypeMatch, info.Match)
	}
	if info.Db > 0 {
		d.Set(pb.NodeType_NodeTypeDb, info.Db)
	}
	if info.Builder > 0 {
		d.Set(pb.NodeType_NodeTypeBuilder, info.Builder)
	}
	if info.Gm > 0 {
		d.Set(pb.NodeType_NodeTypeGm, info.Gm)
	}
	if info.Gate > 0 {
		d.Set(pb.NodeType_NodeTypeGame, info.Gate)
	}
}
