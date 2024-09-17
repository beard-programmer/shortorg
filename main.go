package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/beard-programmer/shortorg/internal/app"
	appLogger "github.com/beard-programmer/shortorg/internal/app/logger"
)

func main() {
	logger, err := appLogger.NewLogger()
	if err != nil {
		panic(err)
	}

	application, err := app.New(context.Background(), logger)
	if err != nil {
		logger.Error("application setup error", err)
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	defer cancel()
	err = application.Serve(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "application serve error", err)
	}

	logger.Warn("program exits")
}
