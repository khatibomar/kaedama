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

	v1 := http.NewServeMux()
	v1.HandleFunc("/health", api.handleHealth)
	v1.HandleFunc("/proxy", api.handleProxy)

	loggingMiddleware := loggingMiddleware(log)
	cors := corsMiddleware(corsOrigins)
	preservePath := preserveOriginalPath()

	mux.Handle("/v1/", preservePath(http.StripPrefix("/v1", cors(loggingMiddleware(v1)))))

	return mux
}
