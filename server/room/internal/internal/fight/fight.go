package fight

import (
	"fmt"
	"go-actor/common/config/repository/skill"
	"go-actor/common/pb"
	"go-actor/framework/actor"
	"go-actor/library/mlog"
	"go-actor/server/room/internal/module/machine"
	"time"
)

// Fight 游戏内存定义 游戏数据分
type Fight struct {
	actor.Actor
	Data     *pb.FightData
	machine  *machine.Machine // 状态机
	isChange bool             // 是否有数据变更
	IsFinish bool             // 平滑关闭
	Timeout  int64
}

// NewRummyGame 初始化游戏对象
func NewFight(data *pb.FightData) *Fight {
	// 初始化 玩家实体集合技能配置
	data.CharacterSkills = make(map[uint32]*pb.Skills, len(data.Characters))
	for _, character := range data.Characters {
		if data.CharacterSkills[character.ID] == nil {
			data.CharacterSkills[character.ID] = &pb.Skills{
				Data: make([]*pb.Skill, len(character.SkillIDs)),
			}
		}

		for _, skillID := range character.SkillIDs {
			data.CharacterSkills[character.ID].Data = append(data.CharacterSkills[character.ID].Data, skill.MGetID(skillID))
		}
	}

	// 初始化 敌人实体集合技能配置
	data.EnemyIntents = make(map[uint32]*pb.Intents, len(data.Ememys))
	for _, ememy := range data.Ememys {
		if data.EnemyIntents[ememy.ID] == nil {
			data.EnemyIntents[ememy.ID] = &pb.Intents{
				Intent: make([]*pb.Skill, len(ememy.Intents)),
			}
		}

		for _, intent := range ememy.Intents {
			emSkill := skill.MGetID(intent.Id)
			emSkill.SetDamage(uint64(intent.Incr))
			data.EnemyIntents[ememy.ID].Intent = append(data.EnemyIntents[ememy.ID].Intent, emSkill)
		}
	}

	ret := &Fight{Data: data}
	nowMs := time.Now().UnixMilli()

	ret.machine = machine.NewMachine(nowMs, pb.GameState_GAME_INIT)
	ret.Actor.Register(ret)
	ret.Actor.SetId(data.FightId)
	ret.Start()
	return ret
}

func (d *Fight) Init() {
	// 初始化状态机
	nowMs := time.Now().UnixMilli()
	handle := d.machine.GetState()
	if handle == nil {
		panic(fmt.Sprintf("Machine状态机未注册状态: %d", d.machine.GetCurState()))
	}
	handle.OnEnter(nowMs, d.machine.GetCurState(), d)
	// 启动定时器
	head := &pb.Head{SendType: pb.SendType_POINT, ActorName: "Fight", FuncName: "OnTick"}
	err := d.RegisterTimer(head, 50*time.Millisecond, -1)
	if err != nil {
		mlog.Debug(head, "register timer err: %v", err)
	}
}

func (d *Fight) Create() {
	// todo 创建战场环境
	mlog.Infof("当前对局玩家单位: %v", d.Data.Characters)
	mlog.Infof("当前对局玩家技能: %v", d.Data.CharacterSkills)
	mlog.Infof("当前对局敌人单位: %v", d.Data.Ememys)
	mlog.Infof("当前对局敌人意图: %v", d.Data.EnemyIntents)
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

func (d *Fight) GetCurStateTTL() int64 {
	switch d.GetCurState() {
	// todo 后续加入玩家控制
	default:
		return 2 * 1000
	}
}

func (d *Fight) FlushExpireTime(nowMs int64) {
	d.Timeout = nowMs + d.GetCurStateTTL()
	d.machine.SetCurStateStartTime(nowMs)
}

func (d *Fight) OnTick() {
	nowMs := time.Now().UnixMilli()
	d.machine.Handle(nowMs, d)
}
