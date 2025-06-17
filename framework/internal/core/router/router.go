package router

import (
	"poker_server/common/pb"
	"sync/atomic"
	"time"
)

type Router struct {
	*pb.Router
	updateTime int64
}

func (d *Router) GetData() *pb.Router {
	return d.Router
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
	flag := false
	switch nodeType {
	case pb.NodeType_NodeTypeGate:
		if !atomic.CompareAndSwapInt32(&d.Gate, nodeId, nodeId) {
			atomic.StoreInt32(&d.Gate, nodeId)
			flag = true
		}
	case pb.NodeType_NodeTypeRoom:
		if !atomic.CompareAndSwapInt32(&d.Room, nodeId, nodeId) {
			atomic.StoreInt32(&d.Room, nodeId)
			flag = true
		}
	case pb.NodeType_NodeTypeMatch:
		if !atomic.CompareAndSwapInt32(&d.Match, nodeId, nodeId) {
			atomic.StoreInt32(&d.Match, nodeId)
			flag = true
		}
	case pb.NodeType_NodeTypeDb:
		if !atomic.CompareAndSwapInt32(&d.Db, nodeId, nodeId) {
			atomic.StoreInt32(&d.Db, nodeId)
			flag = true
		}
	case pb.NodeType_NodeTypeBuilder:
		if !atomic.CompareAndSwapInt32(&d.Builder, nodeId, nodeId) {
			atomic.StoreInt32(&d.Builder, nodeId)
			flag = true
		}
	case pb.NodeType_NodeTypeGm:
		if !atomic.CompareAndSwapInt32(&d.Gm, nodeId, nodeId) {
			atomic.StoreInt32(&d.Gm, nodeId)
			flag = true
		}
	case pb.NodeType_NodeTypeGame:
		if !atomic.CompareAndSwapInt32(&d.Game, nodeId, nodeId) {
			atomic.StoreInt32(&d.Game, nodeId)
			flag = true
		}

	}
	if flag {
		atomic.StoreInt64(&d.updateTime, time.Now().Unix())
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
