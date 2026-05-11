package middleware

import (
	"net/http"

	"github.com/example/banking-service/internal/transport/http/response"

	"github.com/sirupsen/logrus"
)

func Recovery(log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					log.WithFields(logrus.Fields{
						"request_id": GetRequestID(r.Context()),
						"panic":      recovered,
					}).Error("panic recovered")

					response.WriteErrorMessage(w, http.StatusInternalServerError, "internal_error", "internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
