package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type AppLogger = zap.Logger
type SugaredLogger = zap.SugaredLogger

func String(key string, val string) zap.Field {
	return zap.String(key, val)
}

func NewLogger() (*AppLogger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.CallerKey = ""

	return config.Build()
}
