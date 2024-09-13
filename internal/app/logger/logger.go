package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Logger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.CallerKey = ""

	return config.Build()
}
