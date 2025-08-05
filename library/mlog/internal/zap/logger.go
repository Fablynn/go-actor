package zap

import (
	"net/url"
	"os"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 分装zap日志库接口使用
type Logger struct {
	logger *zap.SugaredLogger
	level  int32
}

func init() {
	// 注册winfile协议处理器
	zap.RegisterSink("winfile", func(u *url.URL) (zap.Sink, error) {
		return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	})
}

func NewLogger(env string, level int32, logfile string) interface{} {
	var config zap.Config
	switch env {
	case "release":
		config = zap.NewProductionConfig()
	default:
		config = zap.NewDevelopmentConfig()
	}
	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		//StacktraceKey:  "stacktrace",
	}
	config.OutputPaths = []string{"stdout", logfile}
	config.ErrorOutputPaths = []string{"stderr", logfile}
	// 创建日志
	logger, err := config.Build(zap.AddCaller(), zap.AddCallerSkip(2))
	if err != nil {
		return err
	}
	return &Logger{logger: logger.Sugar(), level: level}
}

func (d *Logger) Close() error {
	if d.logger != nil {
		return d.logger.Sync()
	}
	return nil
}

func (d *Logger) GetLevel() int32 {
	return atomic.LoadInt32(&d.level)
}

func (d *Logger) Debugf(format string, args ...interface{}) {
	d.logger.Debugf(format, args...)
}

func (d *Logger) Warnf(format string, args ...interface{}) {
	d.logger.Warnf(format, args...)
}

func (d *Logger) Infof(format string, args ...interface{}) {
	d.logger.Infof(format, args...)
}

func (d *Logger) Errorf(format string, args ...interface{}) {
	d.logger.Errorf(format, args...)
}

func (d *Logger) Fatalf(format string, args ...interface{}) {
	d.logger.Fatalf(format, args...)
}
