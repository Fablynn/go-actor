package test

import (
	"go-actor/common/config"
	"go-actor/common/config/repository/buffer_config"
	"go-actor/common/config/repository/character"
	"go-actor/common/config/repository/enemys"
	"go-actor/common/config/repository/skill"
	"testing"
)

func TestRpg(t *testing.T) {
	t.Run("ShowAllBuffers", func(t *testing.T) {
		err := config.InitConfig("../../../../gameconf/data", nil)
		t.Logf("err: %v", err)
		buffers := buffer_config.LGet()
		t.Logf("buffers: %v", buffers)
		skills := skill.LGet()
		t.Logf("skills: %v", skills)
		character := character.MGetID(1)
		t.Logf("characters: %v", character.SkillIDs)
		ememy := enemys.MGetID(1)
		t.Logf("ememy indint: %v", ememy.Intents)
	})
}
