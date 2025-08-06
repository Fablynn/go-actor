package test

import (
	"math/bits"
	"testing"
)

// 线段树建设与位运算优化
func TestLineTree(t *testing.T) {
	t.Run("buildLineTree", func(t *testing.T) {
		list := []int{3, 5, 4}
		seg := newSegmentTree(list)
		t.Logf("seg: %v", seg)
		seg.findFirstAndUpdate(1, 0, len(list)-1, 4)
		t.Logf("seg: %v", seg)
	})
}

type segment []int //平衡二叉树

func (s segment) maintain(o int) {
	s[o] = max(s[2*o], s[2*o+1])
}

func (s segment) build(arr []int, o, l, r int) {
	if l == r {
		s[o] = arr[l]
		return
	}
	m := l + r/2
	s.build(arr, 2*o, l, m)
	s.build(arr, 2*o+1, m+1, r)
	s.maintain(o)
}

func newSegmentTree(a []int) segment {
	n := len(a)
	t := make(segment, 2<<bits.Len(uint(n-1)))
	t.build(a, 1, 0, n-1)
	return t
}

// 找区间内的第一个 >= x 的数，并更新为 -1，返回这个数的下标（没有则返回 -1）
func (s segment) findFirstAndUpdate(o, l, r, x int) int {
	if s[o] < x { // 区间没有 >= x 的数
		return -1
	}
	if l == r {
		s[o] = -1 // 更新为 -1，表示不能放水果
		return l
	}
	m := (l + r) / 2
	i := s.findFirstAndUpdate(o*2, l, m, x) // 先递归左子树
	if i < 0 {
		i = s.findFirstAndUpdate(o*2+1, m+1, r, x) // 再递归右子树
	}
	s.maintain(o)
	return i
}
