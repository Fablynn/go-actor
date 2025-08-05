package logger

import (
	"bytes"
	"fmt"
	"go-actor/library/mlog/define"
	"path"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Logger struct {
	level   int32
	writers []IWriter
	pool    sync.Pool
}

func NewLogger(level int32, writers ...IWriter) *Logger {
	return &Logger{
		level:   level,
		writers: writers,
		pool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, 1024))
			},
		},
	}
}

func (l *Logger) Close() error {
	var err error
	for _, writer := range l.writers {
		if e := writer.Close(); e != nil {
			err = e
		}
	}
	return err
}

func (l *Logger) SetLevel(level int32) {
	atomic.StoreInt32(&l.level, level)
}

func (d *Logger) Trace(skip int, format string, args ...interface{}) {
	if atomic.LoadInt32(&d.level) <= define.LOG_TRACE {
		d.output(skip+1, define.LOG_TRACE, format, args...)
	}
}

func (d *Logger) Debug(skip int, format string, args ...interface{}) {
	if atomic.LoadInt32(&d.level) <= define.LOG_DEBUG {
		d.output(skip+1, define.LOG_DEBUG, format, args...)
	}
}

func (d *Logger) Warn(skip int, format string, args ...interface{}) {
	if atomic.LoadInt32(&d.level) <= define.LOG_WARN {
		d.output(skip+1, define.LOG_WARN, format, args...)
	}
}

func (d *Logger) Info(skip int, format string, args ...interface{}) {
	if atomic.LoadInt32(&d.level) <= define.LOG_INFO {
		d.output(skip+1, define.LOG_INFO, format, args...)
	}
}

func (d *Logger) Error(skip int, format string, args ...interface{}) {
	if atomic.LoadInt32(&d.level) <= define.LOG_ERROR {
		d.output(skip+1, define.LOG_ERROR, format, args...)
	}
}

func (d *Logger) Fatal(skip int, format string, args ...interface{}) {
	if atomic.LoadInt32(&d.level) <= define.LOG_FATAL {
		d.output(skip+1, define.LOG_FATAL, format, args...)
	}
}

// 格式化输出
func (d *Logger) output(skip int, level int32, format string, args ...interface{}) {
	tt := time.Now()
	builder := Get(tt)
	builder.WriteByte('[')
	builder.WriteString(tt.Format("2006-01-02 15:04:05.000"))
	builder.WriteString("] [")
	builder.WriteString(define.LevelToString(level))
	builder.WriteString("] ")
	if skip > 0 {
		pc, _, _, _ := runtime.Caller(skip + 1)
		fname := path.Base(runtime.FuncForPC(pc).Name())
		//builder.WriteByte('[')
		//builder.WriteString(file)
		//builder.WriteByte(':')
		//builder.WriteString(strconv.Itoa(line))
		//builder.WriteString("]\t[")
		builder.WriteString(fname)
		builder.WriteByte('\t')
		//builder.WriteString("]\t")
	}
	builder.WriteString(fmt.Sprintf(format, args...))
	builder.WriteByte('\n')

	for _, ww := range d.writers {
		builder.Add(1)
		ww.Write(builder)
	}
}
