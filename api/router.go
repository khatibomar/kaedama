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
) http.Handler {
	api := &api{
		proxyService: proxyService,
		cache:        cache.New(cacheTTL),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", api.handleHealth)
	mux.HandleFunc("/proxy", api.handleProxy)
	mux.HandleFunc("/cache/clear", api.handleCacheClear)

	loggingMiddleware := loggingMiddleware(log)
	cors := corsMiddleware(corsOrigins)
	handler := http.Handler(mux)
	handler = cors(loggingMiddleware(handler))

	return handler
}
