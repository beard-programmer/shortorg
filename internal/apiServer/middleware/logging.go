package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrappedWriter := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrappedWriter, r)

		if 500 <= wrappedWriter.statusCode {
			duration := time.Since(start)
			logger.Info("HTTP request",
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
				zap.String("proto", r.Proto),
				zap.Int("status_code", wrappedWriter.statusCode),
				zap.String("user_agent", r.UserAgent()),
				zap.Float64("duration_ms", duration.Seconds()*1000),
			)
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}
