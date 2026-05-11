package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func Logging(log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now().UTC()

			recorder := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			next.ServeHTTP(recorder, r)

			log.WithFields(logrus.Fields{
				"request_id":  GetRequestID(r.Context()),
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      recorder.status,
				"duration_ms": time.Since(startedAt).Milliseconds(),
			}).Info("http request completed")
		})
	}
}
