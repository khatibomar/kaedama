package api

import (
	"net/http"

	"github.com/khatibomar/kaedama/proxy"
)

type api struct {
	proxyService *proxy.Service
}

func New(proxyService *proxy.Service) http.Handler {
	api := &api{
		proxyService: proxyService,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/proxy", api.handleProxy)

	return mux
}
