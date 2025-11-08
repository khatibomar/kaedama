package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Context key for storing original path
type contextKey string

const originalPathKey contextKey = "originalPath"

// preserveOriginalPath middleware stores the original URL path in context
func preserveOriginalPath() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Store original path in context
			ctx := context.WithValue(r.Context(), originalPathKey, r.URL.Path)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// GetOriginalPath retrieves the original path from context
func GetOriginalPath(r *http.Request) string {
	if path, ok := r.Context().Value(originalPathKey).(string); ok {
		return path
	}
	return r.URL.Path
}

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

func corsMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Set CORS headers
			if allowedOrigins == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				origins := strings.SplitSeq(allowedOrigins, ",")
				for allowedOrigin := range origins {
					allowedOrigin = strings.TrimSpace(allowedOrigin)
					if origin == allowedOrigin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "3600")

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
