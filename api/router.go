package api

import (
	"log/slog"
	"net/http"

	"github.com/khatibomar/kaedama/proxy"
)

type api struct {
	proxyService *proxy.Service
}

func New(
	log *slog.Logger,
	proxyService *proxy.Service,
	corsOrigins string,
) http.Handler {
	api := &api{
		proxyService: proxyService,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", api.handleHealth)
	mux.HandleFunc("/proxy", api.handleProxy)

	loggingMiddleware := loggingMiddleware(log)
	cors := corsMiddleware(corsOrigins)
	handler := http.Handler(mux)
	handler = cors(loggingMiddleware(handler))

	return handler
}
