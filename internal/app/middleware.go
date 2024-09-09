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
		if duration > 1000*time.Millisecond {
			logger.Warnf("%s - - [%s] \"%s %s %s\" %d %d \"%s\" %.3fms",
				r.RemoteAddr,
				time.Now().Format("02/Jan/2006:15:04:05 -0700"),
				r.Method,
				r.URL.String(),
				r.Proto,
				wrappedWriter.statusCode,
				wrappedWriter.responseSize,
				r.UserAgent(),
				duration.Seconds()*1000,
			)
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode   int
	responseSize int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseSize += size
	return size, err
}
