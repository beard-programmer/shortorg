package apiServer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"go.uber.org/zap"
)

func (s *Server) serveHTTP(ctx context.Context) error {

	serverMux, err := s.getServerMux(ctx)
	if err != nil {
		return fmt.Errorf("create http server mux: %w", err)
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", s.config.Host, strconv.Itoa(s.config.HTTP.InternalPort)),
		Handler:      serverMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), GracefulShutdownTimeout)
		defer shutdownCancel()
		s.logger(ctx).Warn("context canceled, shutting down gracefully with timeout",
			zap.Duration("timeout", GracefulShutdownTimeout),
		)

		s.logger(ctx).Info("shutting down http-server")
		err := httpServer.Shutdown(shutdownCtx)
		if err != nil {
			s.logger(ctx).Error("http shutdown error", zap.Error(err))
			<-shutdownCtx.Done()
			s.logger(ctx).Fatal(
				"shutting down timeout reached, stopping through Fatal",
				zap.Duration("timeout", GracefulShutdownTimeout),
			)
		}
		s.logger(ctx).Info("successfully shut down http server")
	}()

	s.logger(ctx).Info("starting http-server", zap.String("addr", httpServer.Addr), zap.Int("concurrency", runtime.GOMAXPROCS(0)))

	err = httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http serve: %w", err)
	}

	return nil
}
