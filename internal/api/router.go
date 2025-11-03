package api

import (
	"net/http"

	"github.com/khatibomar/kaedama/internal/domain/dto"
)

type proxyService interface {
	ProxyURL(requestURL string) (*dto.ProxyResult, error)
}

type api struct {
	proxyService proxyService
}

func New(proxyService proxyService) http.Handler {
	api := &api{
		proxyService: proxyService,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/proxy", api.handleProxy)

	return mux
}
