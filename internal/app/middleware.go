package app

import (
	"bytes"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.SugaredLogger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrappedWriter := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrappedWriter, r)

		if wrappedWriter.statusCode != 200 {
			duration := time.Since(start)
			logger.Errorf("%s - - [%s] \"%s %s %s\" %d %d  \"%s\" %.3fms \nResponse Body: %s",
				r.RemoteAddr,
				time.Now().Format("02/Jan/2006:15:04:05 -0700"),
				r.Method,
				r.URL.String(),
				r.Proto,
				wrappedWriter.statusCode,
				wrappedWriter.responseSize,
				r.UserAgent(),
				duration.Seconds()*1000,
				wrappedWriter.body.String(),
			)
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode   int
	responseSize int
	body         bytes.Buffer
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.body.Write(b) // Capture the response body
	size, err := r.ResponseWriter.Write(b)
	r.responseSize += size
	return size, err
}
