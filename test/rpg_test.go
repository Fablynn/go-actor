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

	t.Run("bug", func(t *testing.T) {
		s := make([]int, 4, 4)
		s[0] = 1
		s[1] = 2
		s[2] = 3
		s[3] = 4
		s1 := upSlice(s)
		s1[0] = 1
		t.Logf("upSlice:%v %v", s, s1)
	})
}

func upSlice(tmp []int) []int {
	tmp[0] = 99
	tmp = append(tmp, 99)
	tmp = append(tmp, 100)
	return tmp
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
