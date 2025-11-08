package api

import (
	"errors"
	"net/http"
	"net/url"

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

	targetURL := urls[0]
	parsedOriginalURL, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	resp, err := api.proxyService.URL(r.Context(), parsedOriginalURL)
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
	for k, v := range resp.Headers {
		if k != "Content-Type" && k != "Content-Length" && k != "Content-Encoding" {
			w.Header().Set(k, v)
		}
	}
	w.WriteHeader(resp.Status)

	// Process M3U8 content if needed
	if proxy.IsM3U8ContentType(resp.ContentType) || proxy.IsM3U8URL(targetURL) {
		// Get the proxy URL for rewriting
		proxyURL := r.URL.Scheme
		if proxyURL == "" {
			proxyURL = "http"
		}
		proxyURL = proxyURL + "://" + r.Host + r.URL.Path

		processedContent := api.proxyService.ProcessM3U8(resp.Content, parsedOriginalURL, proxyURL)
		_, _ = w.Write([]byte(processedContent))
	} else {
		// Write raw content for non-M3U8 responses
		_, _ = w.Write(resp.Content)
	}
}

func (api *api) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(http.StatusText(http.StatusOK)))
}
