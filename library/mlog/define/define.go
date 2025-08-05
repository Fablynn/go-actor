package define

import "strings"

const (
	LOG_TRACE = 1
	LOG_DEBUG = 2
	LOG_WARN  = 3
	LOG_INFO  = 4
	LOG_ERROR = 5
	LOG_FATAL = 6
)

type ILog interface {
	Close() error
	SetLevel(int32)
	Trace(skip int, format string, args ...interface{})
	Debug(skip int, format string, args ...interface{})
	Warn(skip int, format string, args ...interface{})
	Info(skip int, format string, args ...interface{})
	Error(skip int, format string, args ...interface{})
	Fatal(skip int, format string, args ...interface{})
}

func LevelToString(level int32) string {
	switch level {
	case LOG_TRACE:
		return "TRACE"
	case LOG_DEBUG:
		return "DEBUG"
	case LOG_WARN:
		return "WARN"
	case LOG_INFO:
		return "INFO"
	case LOG_ERROR:
		return "ERROR"
	case LOG_FATAL:
		return "FATAL"
	}
	return "DEBUG"
}

func StringToLevel(str string) int32 {
	switch strings.ToUpper(str) {
	case "TRACE":
		return LOG_TRACE
	case "DEBUG":
		return LOG_DEBUG
	case "WARN":
		return LOG_WARN
	case "INFO":
		return LOG_INFO
	case "ERROR":
		return LOG_ERROR
	case "FATAL":
		return LOG_FATAL
	}
	return LOG_WARN
}
