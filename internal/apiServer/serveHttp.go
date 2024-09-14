package apiServer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

func (s *Server) serveHTTP(ctx context.Context) error {
	mux := http.NewServeMux()

	gatewayMux, err := s.getServerMux(ctx)
	if err != nil {
		return fmt.Errorf("create http server mux: %w", err)
	}

	mux.Handle("/", gatewayMux)

	httpServer := &http.Server{
		Addr:         s.config.Host + ":" + strconv.Itoa(s.config.HTTP.InternalPort),
		Handler:      s.wrapWithDefaultMiddlewares(mux),
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
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

	s.logger(ctx).Info("starting http-server", zap.String("addr", httpServer.Addr))

	err = httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http serve: %w", err)
	}

	return nil
}
