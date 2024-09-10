package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/beard-programmer/shortorg/internal"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, shutting down gracefully...")
		cancel()
	}()

	app := new(internal.App).New(ctx)

	if err := app.StartServer(ctx); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	<-ctx.Done()

	gracefulShutdownTimeout := 2 * time.Second
	log.Printf("Waiting %v for graceful shutdown...", gracefulShutdownTimeout)
	time.Sleep(gracefulShutdownTimeout)
	log.Println("Shutdown complete.")
}
