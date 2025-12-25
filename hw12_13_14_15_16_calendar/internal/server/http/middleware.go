package internalhttp

import (
	"fmt"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(logg Logger, next http.Handler) http.Handler { //nolint:unused
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)

		logg.Info(
			fmt.Sprintf(
				"%s [%v] %s %s %s %d %v %v",
				r.RemoteAddr, start.Format("02/Jan/2006:15:04:05 -0700"),
				r.Method, r.URL.Path,
				r.Proto, lrw.statusCode, time.Since(start), r.UserAgent(),
			),
		)
	})
}
