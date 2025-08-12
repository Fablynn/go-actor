package state

import (
	"go-actor/common/pb"
	"go-actor/library/mlog"
	"go-actor/library/random"
	"go-actor/server/room/internal/internal/fight"
	"go-actor/server/room/internal/module/util"
)

/*
	FUNC: // 玩家战斗阶段
*/

type ActState struct {
	BaseState
}

func (d *ActState) OnEnter(nowMs int64, curState pb.GameState, extra interface{}) {
	game := extra.(*fight.Fight)
	game.FlushExpireTime(nowMs)

	for _, character := range game.Data.Characters {
		if character.Status != pb.Status_StatusAlive {
			continue
		}
		skillId := character.SkillIDs[random.Intn(len(character.SkillIDs))]
		skills := game.Data.CharacterSkills[character.ID]
		skill := util.GetSkillById(skillId, skills.Data)

		switch skill.TargetType {
		case pb.TargetType_TargetTypeSINGLEENEMY:
			targets := util.RandAliveEnemy(game.Data.Ememys, skill.TargetCount)
			for _, target := range targets {
				target.Health -= int64(skill.Damage)
				if target.Health <= 0 {
					target.Health = 0
					target.Status = pb.Status_StatusDeath
					util.Active(pb.TriggerType_TriggerTypeNone, skills.Data) // todo 死亡触发技能
					mlog.Infof("玩家操作阶段 %s【%s】击杀：%s(%d/%d)", character.Name, skill.SkillName, target.Name, target.Health, target.MaxHp)
					if len(util.GetAliveEnemy(game.Data.Ememys)) == 0 {
						game.IsFinish = true
						break
					}
				}
				mlog.Infof("玩家操作阶段 %s【%s】：%s(%d/%d)", character.Name, skill.SkillName, target.Name, target.Health, target.MaxHp)
			}
		default:
			mlog.Infof("未编辑技能")
		}
	}
}

func (d *ActState) OnExit(nowMs int64, curState pb.GameState, extra interface{}) {

}
