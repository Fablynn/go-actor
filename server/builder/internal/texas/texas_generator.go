package texas

import (
	"go-actor/common/pb"
	"go-actor/common/room_util"
	"go-actor/framework/actor"
	"go-actor/library/timer"
	"go-actor/library/uerror"
	"reflect"
	"time"
)

type BuilderTexasGenerator struct {
	actor.Actor
	datas    map[pb.GeneratorType]uint32 // 房间ID生成器数据
	isChange bool                        // 是否有变更
	loadFlag uint64                      // 是否加载完成
}

func NewBuilderTexasGenerator() *BuilderTexasGenerator {
	ret := &BuilderTexasGenerator{}
	ret.Actor.Register(ret)
	ret.Actor.ParseFunc(reflect.TypeOf(ret))
	ret.Actor.SetId(uint64(pb.GeneratorType_GeneratorTypeTexas))
	ret.Actor.Start()
	actor.Register(ret)
	return ret
}

func (g *BuilderTexasGenerator) Load() error {
	g.loadFlag = 1
	return timer.Register(&g.loadFlag, func() { g.SendMsg(&pb.Head{FuncName: "LoadRequest"}) }, 5*time.Second, -1)
}

func (g *BuilderTexasGenerator) LoadRequest() error {
	return nil
}

func (g *BuilderTexasGenerator) LoadData(head *pb.Head, rsp *pb.GetGeneratorDataRsp) error {
	if rsp.Head != nil {
		return uerror.ToError(rsp.Head)
	}
	if g.loadFlag > 0 {
		g.loadFlag = 0
		g.datas = make(map[pb.GeneratorType]uint32)
		for _, item := range rsp.List {
			g.datas[item.Id] = item.Incr
		}
		return g.RegisterTimer(&pb.Head{FuncName: "OnTick"}, 5*time.Second, -1)
	}
	return nil
}

func (g *BuilderTexasGenerator) OnTick() {
	if g.isChange {
		g.SendMsg(&pb.Head{FuncName: "Save"})
	}
}

// 保存数据
func (g *BuilderTexasGenerator) Save() error {
	g.isChange = false
	return nil
}

// 生成房间ID请求（同步+异步）
func (g *BuilderTexasGenerator) GenRoomIdReq(head *pb.Head, req *pb.GenRoomIdReq, rsp *pb.GenRoomIdRsp) error {
	if req.Count <= 1 {
		req.Count = 1
	}

	// 随便使用简单方法生成唯一ID
	for i := int32(0); i < req.Count; i++ {
		switch req.GeneratorType {
		case pb.GeneratorType_GeneratorTypeTexas:
			g.datas[req.GeneratorType]++
			incr := g.datas[req.GeneratorType]
			id := room_util.ToTexasRoomId(req.MatchType, req.GameType, req.CoinType) | uint64(incr&0xFFFFFF)
			rsp.RoomIdList = append(rsp.RoomIdList, id)
			g.isChange = true
		default:
			return uerror.New(1, pb.ErrorCode_TYPE_NOT_SUPPORTED, "不支持的生成器类型: %s", req.GeneratorType.String())
		}
	}
	return nil
}
