package api

import (
	"errors"
	"net/http"

	"github.com/khatibomar/kaedama/proxy"
)

func (api *api) handleProxy(w http.ResponseWriter, r *http.Request) {
	urls, ok := r.URL.Query()["url"]
	if !ok {
		http.Error(w, "must provide url", http.StatusBadRequest)
		return
	}

	if len(urls) != 1 {
		http.Error(w, "must contain only one url", http.StatusBadRequest)
		return
	}

	url := urls[0]

	resp, err := api.proxyService.URL(url)
	if err != nil {
		var errValidation *proxy.ValidationError
		if errors.As(err, &errValidation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", resp.ContentType)
}

func (api *api) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(http.StatusText(http.StatusOK)))
}
