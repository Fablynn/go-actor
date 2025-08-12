package state

import (
	"go-actor/common/pb"
	"go-actor/library/mlog"
	"go-actor/server/room/internal/internal/fight"
	"go-actor/server/room/internal/module/util"
)

/*
	FUNC: // 加载战斗环境相关 战场效果
*/

type EnemyState struct {
	BaseState
}

func (d *EnemyState) OnEnter(nowMs int64, curState pb.GameState, extra interface{}) {
	game := extra.(*fight.Fight)
	game.FlushExpireTime(nowMs)

	ememys := util.GetAliveEnemy(game.Data.Ememys)
	for _, ememy := range ememys {
		intent := game.Data.EnemyIntents[ememy.ID]
		skill := intent.Intent[intent.Couser]

		switch skill.TargetType {
		case pb.TargetType_TargetTypeSINGLEENEMY:
			targets := util.RandAliveCharacters(game.Data.Characters, skill.TargetCount)
			for _, target := range targets {
				target.Health -= int64(skill.Damage)
				mlog.Infof("怪物行动阶段 %s【%s】：%s(%d/%d)", ememy.Name, skill.SkillName, target.Name, target.Health, target.MaxHp)
			}
		default:
			mlog.Infof("未编辑技能")
		}

		intent.Couser = (intent.Couser + 1) % uint32(len(intent.Intent))
	}
}

func (d *EnemyState) OnExit(nowMs int64, curState pb.GameState, extra interface{}) {

}
