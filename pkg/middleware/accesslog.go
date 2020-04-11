package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func AccessLog(logger *zap.SugaredLogger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Infow("ACCESS LOG MIDDLEWARE",
			"method", r.Method,
			"remote_addr", r.RemoteAddr,
			"url", r.URL.Path,
		)
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Infow("REQUEST PROCESSED",
			"time", time.Since(start),
		)
	})
}
