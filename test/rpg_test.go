package test

import (
	"fmt"
	"go-actor/common/config"
	"go-actor/common/config/repository/buffer_config"
	"go-actor/common/config/repository/skill"
	"sync"
	"testing"
	"time"
)

func TestRpg(t *testing.T) {
	t.Run("ShowAllBuffers", func(t *testing.T) {
		config.InitConfig("./gameconf/data", nil)
		buffers := buffer_config.LGet()
		t.Logf("buffers: %v", buffers)
		skills := skill.LGet()
		t.Logf("skills: %v", skills)
	})

	t.Run("testRwDebule", func(t *testing.T) {
		go b()
		go a()
		go c()
		time.Sleep(5 * time.Second)
	})
}

var rw sync.RWMutex

// 写之前读2
func a() {
	rw.RLock()
	defer rw.RUnlock()

	time.Sleep(1 * time.Second)
	fmt.Println("aaaa")
}

func b() {
	rw.RLock()
	defer rw.RUnlock()
	time.Sleep(1 * time.Second)
	a()
	fmt.Println("bbbb")
}

func c() {
	time.Sleep(1 * time.Second)
	rw.Lock()
	defer rw.Unlock()
	fmt.Println("cccc")
}
