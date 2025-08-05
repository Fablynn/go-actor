package test

import (
	"go-actor/library/mlog"
	"testing"
)

func TestMain(m *testing.M) {
	mlog.Init("mlog", 1, "trace", "./log")
	m.Run()
}

func BenchmarkCluster(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mlog.Errorf("----fadfasdfad--skdfakdjfakjdfasdfasdfasdfadfasdfadf----skdfakdjfakjdfasdfasdfasdfadfasdfadfff")
	}
	b.Log(b.N)
}
