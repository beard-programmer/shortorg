package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/beard-programmer/shortorg/internal/app"
	"github.com/beard-programmer/shortorg/internal/app/logger"
	"go.uber.org/zap"
)

func main() {
	zapLogger, err := logger.Logger()
	if err != nil {
		panic(err)
	}

	application, err := app.New(context.Background(), zapLogger)
	if err != nil {
		zapLogger.Fatal("application setup error", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	defer cancel()
	err = application.Serve(ctx)
	if err != nil {
		zapLogger.Error("application serve error", zap.Error(err))
	}

	zapLogger.Warn("program exits")
}
