package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger
var once sync.Once

func InitLogger() {
	once.Do(func() {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "time"
		cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		cfg.OutputPaths = []string{"stdout"}
		l, _ := cfg.Build()
		logger = l
	})
}

func Warn(msg string, fields ...zapcore.Field) {
	logger.Warn(
		msg,
		fields...,
	)
}

func Info(msg string, fields ...zapcore.Field) {
	logger.Info(
		msg,
		fields...,
	)
}

func Error(msg string, fields ...zapcore.Field) {
	logger.Error(
		msg,
		fields...,
	)
}
