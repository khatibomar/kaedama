package api

import (
	"errors"
	"net/http"

	"github.com/khatibomar/kaedama/proxy"
)

func (api *api) handleProxy(rw http.ResponseWriter, req *http.Request) {
	urls, ok := req.URL.Query()["url"]
	if !ok {
		http.Error(rw, "must provide url", http.StatusBadRequest)
		return
	}

	if len(urls) != 1 {
		http.Error(rw, "must contain only one url", http.StatusBadRequest)
		return
	}

	url := urls[0]

	resp, err := api.proxyService.URL(url)
	if err != nil {
		var errValidation *proxy.ValidationError
		if errors.As(err, &errValidation) {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", resp.ContentType)
}
