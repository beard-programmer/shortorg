package app

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.SugaredLogger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrappedWriter := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start)
		logger.Infof("%s - - \"%s %s %s\" %d  \"%s\" %.3fms",
			r.RemoteAddr,
			r.Method,
			r.URL.String(),
			r.Proto,
			wrappedWriter.statusCode,
			r.UserAgent(),
			duration.Seconds()*1000,
		)
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
