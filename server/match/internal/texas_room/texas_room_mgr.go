package texas_room

import (
	"go-actor/common/config/repository/texas_config"
	"go-actor/common/pb"
	"go-actor/common/room_util"
	"go-actor/framework/actor"
	"go-actor/library/mlog"
	"reflect"
)

type MatchTexasRoomMgr struct {
	actor.Actor
	mgr   *actor.ActorMgr
	datas map[uint64]*pb.TexasRoomData
}

func NewMatchTexasRoomMgr() *MatchTexasRoomMgr {
	mgr := new(actor.ActorMgr)
	rr := &MatchTexasRoom{}
	mgr.Register(rr)
	mgr.ParseFunc(reflect.TypeOf(rr))
	actor.Register(mgr)

	ret := &MatchTexasRoomMgr{mgr: mgr}
	ret.Actor.Register(ret)
	ret.Actor.ParseFunc(reflect.TypeOf(ret))
	ret.Actor.SetId(uint64(pb.DataType_DataTypeTexasRoom))
	ret.Actor.Start()
	actor.Register(ret)
	return ret
}

func (m *MatchTexasRoomMgr) Close() {
	m.mgr.Stop()
	m.Actor.Stop()
	mlog.Infof("MatchTexasRoomMgr关闭成功")
}

// Load todo 异步加载房间数据
func (m *MatchTexasRoomMgr) Load() error {
	tmps := map[uint64][]*pb.TexasRoomData{}
	texas_config.Walk(func(cfg *pb.TexasConfig) bool {
		id := room_util.ToMatchGameId(cfg.MatchType, cfg.GameType, cfg.CoinType)
		if _, ok := tmps[id]; ok {
			return true
		}
		m.mgr.AddActor(NewMatchTexasRoom(id))
		return true
	})
	m.datas = make(map[uint64]*pb.TexasRoomData)
	return nil
}

// LoadComplete todo 异步加载成功数据打到对应actor

// OnTick 定时落地到db服务
func (m *MatchTexasRoomMgr) OnTick() {
	if len(m.datas) <= 0 {
		return
	}
	m.SendMsg(&pb.Head{FuncName: "Save"})
}

func (m *MatchTexasRoomMgr) Collect(notify *pb.UpdateTexasRoomDataNotify) {
	for _, item := range notify.List {
		m.datas[item.RoomId] = item
	}
}

// Save 保存数据
func (m *MatchTexasRoomMgr) Save() error {
	if len(m.datas) <= 0 {
		return nil
	}

	//todo dbsvr save data to restart load

	// 清空数据
	for key := range m.datas {
		delete(m.datas, key)
	}
	return nil
}
