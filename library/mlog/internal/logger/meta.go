package logger

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	pool = sync.Pool{
		New: func() interface{} {
			return &MetaData{Buffer: bytes.NewBuffer(nil)}
		},
	}
)

type IWriter interface {
	Write(*MetaData) error
	Close() error
}

func Get(tt time.Time) *MetaData {
	meta := pool.Get().(*MetaData)
	if meta.cb == nil {
		meta.cb = pool.Put
	}
	atomic.StoreInt32(&meta.reference, 0)
	meta.tt = tt
	meta.Reset()
	return meta
}

type MetaData struct {
	*bytes.Buffer
	tt        time.Time
	reference int32
	cb        func(interface{})
}

func (m *MetaData) Add(val int32) int32 {
	return atomic.AddInt32(&m.reference, val)
}

func (m *MetaData) Done() {
	newVal := atomic.AddInt32(&m.reference, -1)
	if newVal == 0 {
		m.cb(m)
	}
}

func (m *MetaData) GetFileName(lname string) string {
	return fmt.Sprintf("%04d%02d%02d/%s-%02d.log", m.tt.Year(), m.tt.Month(), m.tt.Day(), lname, m.tt.Hour())
}
