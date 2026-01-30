package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type logger struct {
	*zap.SugaredLogger
}

var (
	globalLogger *logger
	once         sync.Once
)

func new(env string) *logger {

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	lumberJackLogger := &lumberjack.Logger{
		Filename:   "logs/system.log",
		MaxSize:    5,
		MaxBackups: 10,
		MaxAge:     15,
		Compress:   true,
	}

	var logLevel zapcore.Level

	if env == "development" || env == "local" {
		logLevel = zap.DebugLevel
	} else {
		logLevel = zap.InfoLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(lumberJackLogger), logLevel),
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(os.Stdout), logLevel),
	)

	base := zap.New(core)

	return &logger{base.Sugar()}
}

func InitGlobal(env string) *logger {
	once.Do(func() {
		globalLogger = new(env)
	})
	return globalLogger
}

func Get() *logger {

	if globalLogger == nil {
		InitGlobal(os.Getenv("APP_ENV"))
	}
	return globalLogger
}

func Info(args ...interface{}) {
	Get().Info(args...)
}

func Infof(template string, args ...interface{}) {
	Get().Infof(template, args...)
}

func Error(args ...interface{}) {
	Get().Error(args...)
}

func Errorf(template string, args ...interface{}) {
	Get().Errorf(template, args...)
}

func Warn(args ...interface{}) {
	Get().Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	Get().Warnf(template, args...)
}

func Debug(args ...interface{}) {
	Get().Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	Get().Debugf(template, args...)
}

func Sync() {
	if globalLogger != nil {
		globalLogger.Sync()
	}
}
