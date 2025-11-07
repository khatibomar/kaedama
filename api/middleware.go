package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func loggingMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			log.InfoContext(r.Context(),
				"Request started",
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.String("host", r.Host),
				slog.String("remote", r.RemoteAddr),
			)

			lrw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(lrw, r)

			latency := humanDuration(time.Since(start))
			log.LogAttrs(r.Context(),
				slog.LevelInfo,
				"Request completed",
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.Int("status", lrw.status),
				slog.Int("size", lrw.size),
				slog.String("latency", latency),
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

// humanDuration turns a duration into something like "532µs", "23ms", "2.3s", or "1m12s".
func humanDuration(d time.Duration) string {
	us := d.Microseconds()
	switch {
	case us < 1000:
		return fmt.Sprintf("%dµs", us)
	case us < 1_000_000:
		return fmt.Sprintf("%.2fms", float64(us)/1000)
	case us < 60_000_000:
		return fmt.Sprintf("%.2fs", float64(us)/1_000_000)
	default:
		minute := int(us / 60_000_000)
		sec := float64(us%60_000_000) / 1_000_000
		return fmt.Sprintf("%dm%.1fs", minute, sec)
	}
}
