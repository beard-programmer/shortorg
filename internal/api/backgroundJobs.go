package api

import (
	"context"
)

func (s *Server) serveBackgroundJobs(ctx context.Context) {
	encodeURLChan := s.urlWasEncodedHandler(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				s.logger.WarnContext(
					ctx,
					"context canceled, shutting down background workers",
					"timeout", gracefulShutdownTimeout,
				)

				return
			case err, ok := <-encodeURLChan:
				if !ok {
					s.logger.ErrorContext(ctx, "error channel is closes for worker", err)

					return
				}
				if err != nil {
					s.logger.ErrorContext(ctx, "error url was encoded worker", err)
				}
			}
		}
	}()
}
