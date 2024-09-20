package api

import (
	"context"
	"sync"
	"time"

	appLogger "github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/beard-programmer/shortorg/internal/encode"
	"github.com/beard-programmer/shortorg/internal/resolveLink"
)

const (
	gracefulShutdownTimeout = 5 * time.Second
)

type Server struct {
	encodeFn             encode.Fn
	decodeFn             resolveLink.ResolveLinkFn
	urlWasEncodedHandler encode.SaveEncodedURLJob
	config               Config

	serverName string
	env        string
	logger     *appLogger.AppLogger
}

func New(
	encodeFn encode.Fn,
	decodeFn resolveLink.ResolveLinkFn,
	urlWasEncodedHandler encode.SaveEncodedURLJob,
	logger *appLogger.AppLogger,
	config Config,
	serverName string,
	env string,
) *Server {
	return &Server{
		encodeFn:             encodeFn,
		decodeFn:             decodeFn,
		urlWasEncodedHandler: urlWasEncodedHandler,
		config:               config,
		serverName:           serverName,
		logger:               logger,
		env:                  env,
	}
}

func (s *Server) Serve(ctx context.Context) error {
	serveWg := new(sync.WaitGroup)
	serveWg.Add(1)
	go func() {
		defer serveWg.Done()
		s.serveBackgroundJobs(ctx)
	}()

	serveWg.Add(1)
	var serveHTTPErr error
	go func() {
		defer serveWg.Done()
		serveHTTPErr = s.serveHTTP(ctx)
	}()

	serveWg.Wait()

	if serveHTTPErr != nil {
		return serveHTTPErr
	}

	return nil
}
