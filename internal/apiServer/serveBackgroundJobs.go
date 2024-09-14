package apiServer

import (
	"context"

	"go.uber.org/zap"
)

func (s *Server) serveBackgroundJobs(ctx context.Context) {
	encodeUrlChan := s.urlWasEncodedHandler(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				s.logger(ctx).Warn("context canceled, shutting down background workers",
					zap.Duration("timeout", GracefulShutdownTimeout),
				)
				return
			case err, ok := <-encodeUrlChan:
				if !ok {
					s.logger(ctx).Error("error channel is closes for worker", zap.Error(err))
					return
				}
				if err != nil {
					s.logger(ctx).Error("error url was encoded worker", zap.Error(err))
				}
			}
		}
	}()
}
