package main

import (
	"context"

	"github.com/beard-programmer/shortorg/internal/app"
	"github.com/beard-programmer/shortorg/internal/app/logger"
	"go.uber.org/zap"
)

func main() {
	zapLogger, err := logger.Logger()
	if err != nil {
		panic(err) // Логгера нет, фаталить стандартным не очень хорошо
	}

	application := app.New(zapLogger)

	err = application.Setup(context.Background())
	if err != nil {
		zapLogger.Fatal("application setup error", zap.Error(err))
	}

	err = application.Serve(context.Background())
	if err != nil {
		zapLogger.Error("application serve error", zap.Error(err))
	}

	zapLogger.Warn("program exits")
	//
	//ctx, cancel := context.WithCancel(context.Background())
	//
	//sigChan := make(chan os.Signal, 1)
	//signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	//
	//go func() {
	//	<-sigChan
	//	log.Println("Received shutdown signal, shutting down gracefully...")
	//	cancel()
	//}()
	//
	//app := new(internal.App).New(ctx)
	//
	//if err := app.StartServer(ctx); err != nil {
	//	log.Fatalf("Server failed to start: %v", err)
	//}
	//
	//<-ctx.Done()
	//
	//gracefulShutdownTimeout := 2 * time.Second
	//log.Printf("Waiting %v for graceful shutdown...", gracefulShutdownTimeout)
	//time.Sleep(gracefulShutdownTimeout)
	//log.Println("Shutdown complete.")
}
