package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/beard-programmer/shortorg/internal/decode"
	"github.com/beard-programmer/shortorg/internal/encode"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
)

const (
	httpTimeout = 5 * time.Second
)

func (s *Server) serveHTTP(ctx context.Context) error {
	serverMux := s.getServerMux(ctx)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", s.config.Host, strconv.Itoa(s.config.HTTP.InternalPort)),
		Handler:           serverMux,
		ReadHeaderTimeout: httpTimeout,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer shutdownCancel()
		s.logger.WarnContext(
			ctx,
			"context canceled, shutting down gracefully with timeout",
			"timeout", gracefulShutdownTimeout,
		)

		s.logger.WarnContext(ctx, "shutting down http-server")
		shutdownErr := httpServer.Shutdown(shutdownCtx) //nolint:contextcheck // gracefully shutdown
		if shutdownErr != nil {
			s.logger.ErrorContext(ctx, "http shutdown error", shutdownErr)

			<-shutdownCtx.Done()

			s.logger.ErrorContext(
				ctx,
				"shutting down timeout reached, stopping through panic",
				"timeout",
				gracefulShutdownTimeout,
			)
			panic(shutdownErr)
		}
		s.logger.WarnContext(ctx, "http shutdown complete")
	}()

	s.logger.InfoContext(ctx, "http server started", "addr", httpServer.Addr, "concurrency", runtime.GOMAXPROCS(0))

	err := httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http serve: %w", err)
	}

	return nil
}

func (s *Server) getServerMux(_ context.Context) *chi.Mux {
	mux := s.wrapWithDefaultMiddlewares(chi.NewMux())

	mux.Route(
		"/api", func(r chi.Router) {
			r.Use(middleware.AllowContentType("application/json"))
			r.Post("/encode", encode.HttpHandlerFunc(s.logger, s.encodeFn))
			r.Post("/decode", decode.HTTPHandlerFunc(s.logger, s.decodeFn))
		},
	)

	return mux
}

func (s *Server) wrapWithDefaultMiddlewares(mux *chi.Mux) *chi.Mux {

	logger := httplog.NewLogger(
		"", httplog.Options{
			LogLevel:        slog.LevelDebug,
			Concise:         true,
			RequestHeaders:  true,
			TimeFieldFormat: time.DateTime,
			Tags: map[string]string{
				"env": s.env,
			},
			QuietDownRoutes: []string{
				"/api/decode",
				"/api/encode",
			},
			QuietDownPeriod: 1 * time.Second,
		},
	)
	mux.Use(httplog.RequestLogger(logger, []string{"/ping", "/debug"}))
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Heartbeat("/ping"))
	mux.Use(middleware.Timeout(httpTimeout))
	mux.Mount("/debug", middleware.Profiler())

	return mux
}
