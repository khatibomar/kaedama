package api

import (
	"log/slog"
	"net/http"
	"time"
)

func loggingMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			log.DebugContext(r.Context(),
				"Request started",
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.String("host", r.Host),
				slog.String("remote", r.RemoteAddr),
			)

			lrw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(lrw, r)

			log.DebugContext(r.Context(),
				"Request completed",
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.Int("status", lrw.status),
				slog.Int("size", lrw.size),
				slog.String("latency", time.Since(start).String()),
			)
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := lrw.ResponseWriter.Write(b)
	lrw.size += n
	return n, err
}
