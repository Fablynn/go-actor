package gen

import (
	"go-actor/common/pb"
	"go-actor/common/redis/repository/generator_data"
	"go-actor/framework/actor"
	"go-actor/library/mlog"
	"reflect"
	"time"
)

type Generator struct {
	actor.Actor
	data     map[pb.GeneratorType]*pb.GeneratorData
	isChange bool
}

func NewGenerator() *Generator {
	ret := &Generator{}
	ret.Actor.Register(ret)
	ret.Actor.ParseFunc(reflect.TypeOf(ret))
	ret.Actor.Start()
	actor.Register(ret)
	return ret
}

func (g *Generator) Init() error {
	rets, err := generator_data.HGetAll()
	if err != nil {
		return err
	}

	g.data = make(map[pb.GeneratorType]*pb.GeneratorData)
	for _, item := range rets {
		g.data[item.Id] = item
	}
	return g.RegisterTimer(&pb.Head{FuncName: "OnTick"}, 3*time.Second, -1)
}

func (g *Generator) OnTick() {
	if !g.isChange {
		return
	}
	g.SendMsg(&pb.Head{FuncName: "Save"})
}

func (g *Generator) Close() error {
	g.Save()
	g.Actor.Stop()
	return nil
}

// 保存数据
func (g *Generator) Save() {
	if !g.isChange {
		return
	}
	tmps := map[string]*pb.GeneratorData{}
	for _, item := range g.data {
		tmps[generator_data.GetField(item.Id)] = item
	}
	if err := generator_data.HMSet(tmps); err != nil {
		mlog.Errorf("保存数据失败：%v", err)
	} else {
		g.isChange = false
	}
}

// 生成房间ID请求（同步+异步）
func (g *Generator) GenRoomIdReq(head *pb.Head, req *pb.GenRoomIdReq, rsp *pb.GenRoomIdRsp) error {
	return nil
}
