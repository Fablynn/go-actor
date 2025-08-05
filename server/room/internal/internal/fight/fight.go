package fight

import (
	"go-actor/common/pb"
	"go-actor/framework/actor"
	"go-actor/library/mlog"
	"go-actor/server/room/internal/module/machine"
	"time"
)

// Fight 游戏内存定义 游戏数据分
type Fight struct {
	actor.Actor
	// 状态机
	Data     *pb.FightData
	machine  *machine.Machine //状态机
	isChange bool             // 是否有数据变更
	IsFinish bool             // 平滑关闭
}

// NewRummyGame 初始化游戏对象
func NewFight(data *pb.FightData) *Fight {
	// todo init nil obj

	ret := &Fight{Data: data}
	nowMs := time.Now().UnixMilli()

	ret.machine = machine.NewMachine(nowMs, pb.GameState_GAME_INIT, ret)
	ret.Actor.Register(ret)
	ret.Actor.SetId(data.FightId)
	ret.Start()

	return ret
}

func (d *Fight) Init() {
	// 启动定时器
	head := &pb.Head{SendType: pb.SendType_POINT, ActorName: "RummyGame", FuncName: "OnTick"}
	err := d.RegisterTimer(head, 50*time.Millisecond, -1)
	if err != nil {
		mlog.Debug(head, "register timer err: %v", err)
	}
}

func (d *Fight) Create() {
	// todo 幂等创建战场环境
}

func (d *Fight) GetCurState() pb.GameState {
	return d.machine.GetCurState()
}

func (d *Fight) GetNextState() pb.GameState {
	switch d.GetCurState() {
	case pb.GameState_GAME_INIT:
		return pb.GameState_GAME_LOAD
	case pb.GameState_GAME_LOAD:
		return pb.GameState_GAME_PRE
	case pb.GameState_GAME_PRE:
		return pb.GameState_GAME_ACT
	case pb.GameState_GAME_ACT:
		return pb.GameState_GAME_AFTER
	case pb.GameState_GAME_AFTER:
		return pb.GameState_GAME_ENEMY
	case pb.GameState_GAME_ENEMY:
		return pb.GameState_GAME_PRE
	case pb.GameState_GAME_SETTLE:
		return pb.GameState_GAME_Finish
	}
	return pb.GameState_GAME_INIT
}
