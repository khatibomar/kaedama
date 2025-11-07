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
) http.Handler {
	api := &api{
		proxyService: proxyService,
	}

	mux := http.NewServeMux()

	v1 := http.NewServeMux()
	v1.HandleFunc("/health", api.handleHealth)
	v1.HandleFunc("/proxy", api.handleProxy)

	loggingMiddleware := loggingMiddleware(log)

	mux.Handle("/v1/", http.StripPrefix("/v1", loggingMiddleware(v1)))

	return mux
}
