package random

import (
	"math/rand"
	"time"
)

var randObj = rand.New(rand.NewSource(time.Now().UnixNano()))

// [0,n)
func Intn(n int) int {
	return randObj.Intn(n)
}

// [0,n)
func Uint32n(n uint32) uint32 {
	return uint32(Intn(int(n)))
}

// [0,n)
func Int32n(n int32) int32 {
	return int32(Intn(int(n)))
}

// [0,n)
func Int64n(n int64) int64 {
	return randObj.Int63n(n)
}

// 取一个范围的随机数
// [2, 5]
func Uint32Part(min, max uint32) uint32 {
	// [2, 5]
	n := max - min + 1
	return uint32(randObj.Int31n(int32(n))) + min
}

func Int64Part(min, max int64) int64 {
	// [2, 5]
	n := max - min + 1
	return randObj.Int63n(n) + min
}

func Int32Part(min, max int32) int32 {
	// [2, 5]
	n := max - min + 1
	return randObj.Int31n(n) + min
}
