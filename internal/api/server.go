package api

import (
	"context"
	"sync"
	"time"

	"github.com/beard-programmer/shortorg/internal/decode"
	"github.com/beard-programmer/shortorg/internal/encode"
	"github.com/beard-programmer/shortorg/internal/encode/infrastructure"
	"go.uber.org/zap"
)

const (
	gracefulShutdownTimeout = 5 * time.Second
)

type Server struct {
	encodeFn             encode.Fn
	decodeFn             decode.Fn
	urlWasEncodedHandler infrastructure.URLWasEncodedHandlerFn
	config               Config

	serverName string
	_logger    *zap.Logger
}

func New(
	encodeFn encode.Fn,
	decodeFn decode.Fn,
	urlWasEncodedHandler infrastructure.URLWasEncodedHandlerFn,
	zapLogger *zap.Logger,
	config Config,
	serverName string,
) *Server {
	return &Server{
		encodeFn:             encodeFn,
		decodeFn:             decodeFn,
		urlWasEncodedHandler: urlWasEncodedHandler,
		config:               config,
		serverName:           serverName,
		_logger:              zapLogger,
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

func (s *Server) logger(_ context.Context) *zap.Logger {
	// TODO: add context to logger
	return s._logger
}
