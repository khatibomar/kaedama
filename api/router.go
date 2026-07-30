package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/khatibomar/kaedama/cache"
	"github.com/khatibomar/kaedama/proxy"
)

type api struct {
	proxyService *proxy.Service
	cache        *cache.Cache
}

func New(
	log *slog.Logger,
	proxyService *proxy.Service,
	corsOrigins string,
	cacheTTL time.Duration,
	maxCacheSize int64,
) http.Handler {
	api := &api{
		proxyService: proxyService,
		cache:        cache.New(cacheTTL, maxCacheSize),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", api.handleHealth)
	mux.HandleFunc("GET /proxy", api.handleProxy)
	mux.HandleFunc("GET /cache/clear", api.handleCacheClear)

	loggingMiddleware := loggingMiddleware(log)
	cors := corsMiddleware(corsOrigins)
	handler := http.Handler(mux)
	handler = cors(loggingMiddleware(handler))

	return handler
}
