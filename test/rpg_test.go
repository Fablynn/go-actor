package test

import (
	"go-actor/common/config"
	"go-actor/common/config/repository/buffer_config"
	"go-actor/common/config/repository/skill_config"
	"testing"
)

func TestRpg(t *testing.T) {
	t.Run("ShowAllBuffers", func(t *testing.T) {
		config.InitConfig("./gameconf/data", nil)
		buffers := buffer_config.LGet()
		t.Logf("buffers: %v", buffers)
		skills := skill_config.LGet()
		t.Logf("skills: %v", skills)
	})
}
