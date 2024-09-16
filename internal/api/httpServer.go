package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/beard-programmer/shortorg/internal/decode"
	"github.com/beard-programmer/shortorg/internal/encode"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func (s *Server) serveHTTP(ctx context.Context) error {
	const (
		readTimeout  = 5 * time.Second
		writeTimeout = 5 * time.Second
	)
	serverMux, err := s.getServerMux(ctx)
	if err != nil {
		return fmt.Errorf("create http server mux: %w", err)
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", s.config.Host, strconv.Itoa(s.config.HTTP.InternalPort)),
		Handler:      serverMux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer shutdownCancel()
		s.logger(ctx).Warn(
			"context canceled, shutting down gracefully with timeout",
			zap.Duration("timeout", gracefulShutdownTimeout),
		)

		s.logger(ctx).Info("shutting down http-server")
		shutdownErr := httpServer.Shutdown(shutdownCtx)
		if shutdownErr != nil {
			s.logger(ctx).Error("http shutdown error", zap.Error(shutdownErr))
			<-shutdownCtx.Done()
			s.logger(ctx).Fatal(
				"shutting down timeout reached, stopping through Fatal",
				zap.Duration("timeout", gracefulShutdownTimeout),
			)
		}
		s.logger(ctx).Info("successfully shut down http server")
	}()

	s.logger(ctx).Info(
		"starting http-server",
		zap.String("addr", httpServer.Addr),
		zap.Int("concurrency", runtime.GOMAXPROCS(0)),
	)

	err = httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http serve: %w", err)
	}

	return nil
}

func (s *Server) getServerMux(ctx context.Context) (*chi.Mux, error) {
	mux := s.wrapWithDefaultMiddlewares(chi.NewMux())

	mux.Route(
		"/api", func(r chi.Router) {
			r.Use(middleware.AllowContentType("application/json"))
			r.Post("/encode", encode.HttpHandlerFunc(s.logger(ctx), s.encodeFn))
			r.Post("/decode", decode.HttpHandlerFunc(s.logger(ctx), s.decodeFn))
		},
	)
	return mux, nil
}

func (s *Server) wrapWithDefaultMiddlewares(mux *chi.Mux) *chi.Mux {
	mux.Use(middleware.Logger)
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Heartbeat("/ping"))
	mux.Mount("/debug", middleware.Profiler())

	return mux
}
