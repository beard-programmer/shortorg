package apiServer

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
	GracefulShutdownTimeout = 5 * time.Second
)

type Server struct {
	encodeFn             encode.Fn
	decodeFn             decode.Fn
	urlWasEncodedHandler infrastructure.UrlWasEncodedHandlerFn
	config               Config

	serverName string
	_logger    *zap.Logger
}

func New(
	encodeFn encode.Fn,
	decodeFn decode.Fn,
	urlWasEncodedHandler infrastructure.UrlWasEncodedHandlerFn,
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
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.serveBackgroundJobs(ctx)
	}()

	wg.Add(1)
	var serveHTTPErr error
	go func() {
		defer wg.Done()
		serveHTTPErr = s.serveHTTP(ctx)
	}()

	wg.Wait()

	if serveHTTPErr != nil {
		return serveHTTPErr
	}

	return nil
}

func (s *Server) logger(ctx context.Context) *zap.Logger {
	// TODO: add context to logger
	return s._logger
}
