package mlog

import (
	"fmt"
	"go-actor/common/pb"
	"go-actor/library/mlog/define"
	"go-actor/library/mlog/internal/filter"
	"go-actor/library/mlog/internal/logger"
	"strings"
)

var obj define.ILog = logger.NewLogger(define.LOG_DEBUG, &logger.StdWriter{})

func Init(appname string, id int32, level, logpath string) {
	obj.Close()
	logname := fmt.Sprintf("%s%d", strings.ToLower(appname), id)
	loglevel := define.StringToLevel(level)
	obj = logger.NewLogger(loglevel, logger.NewLogWriter(logpath, logname, 1024), &logger.StdWriter{})
}

func Close() error {
	return obj.Close()
}

func SetLevel(level int32) {
	obj.SetLevel(level)
}

func Trace(head *pb.Head, format string, args ...interface{}) {
	if !filter.IsFilter(head) {
		obj.Trace(1, filter.Filter(head, format), args...)
	}
}

func Debug(head *pb.Head, format string, args ...interface{}) {
	if !filter.IsFilter(head) {
		obj.Debug(1, filter.Filter(head, format), args...)
	}
}

func Warn(head *pb.Head, format string, args ...interface{}) {
	if !filter.IsFilter(head) {
		obj.Warn(1, filter.Filter(head, format), args...)
	}
}

func Info(head *pb.Head, format string, args ...interface{}) {
	if !filter.IsFilter(head) {
		obj.Info(1, filter.Filter(head, format), args...)
	}
}

func Error(head *pb.Head, format string, args ...interface{}) {
	if !filter.IsFilter(head) {
		obj.Error(-1, filter.Filter(head, format), args...)
	}
}

func Fatal(head *pb.Head, format string, args ...interface{}) {
	if !filter.IsFilter(head) {
		obj.Fatal(1, filter.Filter(head, format), args...)
	}
}

// --------Debugf---------
func Tracef(format string, args ...interface{}) {
	obj.Trace(1, format, args...)
}

func Debugf(format string, args ...interface{}) {
	obj.Debug(1, format, args...)
}

func Warnf(format string, args ...interface{}) {
	obj.Warn(1, format, args...)
}

func Infof(format string, args ...interface{}) {
	obj.Info(1, format, args...)
}

func Errorf(format string, args ...interface{}) {
	obj.Error(1, format, args...)
}

func Fatalf(format string, args ...interface{}) {
	obj.Fatal(1, format, args...)
}
