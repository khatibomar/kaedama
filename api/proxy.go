package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/khatibomar/kaedama/proxy"
)

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	ContentType string
	Headers     map[string]string
	Status      int
	Body        []byte
}

// remoteHeaderDenylist contains headers that the remote host must not be
// allowed to set on our responses. These are managed exclusively by our own
// middleware (e.g. corsMiddleware).
var remoteHeaderDenylist = map[string]struct{}{
	"Access-Control-Allow-Origin":      {},
	"Access-Control-Allow-Methods":     {},
	"Access-Control-Allow-Headers":     {},
	"Access-Control-Allow-Credentials": {},
	"Access-Control-Max-Age":           {},
	"Access-Control-Expose-Headers":    {},
	"Content-Type":                     {},
	"Content-Length":                   {},
	"Content-Encoding":                 {},
}

// stripRemoteHeaders copies headers from a remote response map into w,
// skipping any header that we manage ourselves.
func stripRemoteHeaders(w http.ResponseWriter, headers map[string]string) {
	for k, v := range headers {
		if _, denied := remoteHeaderDenylist[k]; !denied {
			w.Header().Set(k, v)
		}
	}
}

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

	if cached, exists := api.cache.Get(targetURL); exists {
		if cachedResp, ok := cached.(*CachedResponse); ok {
			w.Header().Set("Content-Type", cachedResp.ContentType)
			stripRemoteHeaders(w, cachedResp.Headers)
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(cachedResp.Status)
			_, _ = w.Write(cachedResp.Body) //nolint:gosec // G705: proxy passthrough by design
			return
		}
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
	stripRemoteHeaders(w, resp.Headers)
	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(resp.Status)

	var responseBody []byte

	// Process M3U8 content if needed
	if (proxy.IsM3U8ContentType(resp.ContentType) || proxy.IsM3U8URL(targetURL)) && proxy.IsActualM3U8Content(resp.Content) {
		proxyURI := "/proxy"
		processedContent := api.proxyService.ProcessM3U8(resp.Content, parsedOriginalURL, proxyURI)
		responseBody = []byte(processedContent)
	} else {
		// Use raw content for non-M3U8 responses
		responseBody = resp.Content
	}

	// Cache successful responses (2xx status codes)
	if resp.Status >= 200 && resp.Status < 300 {
		cachedResp := &CachedResponse{
			ContentType: resp.ContentType,
			Headers:     resp.Headers,
			Status:      resp.Status,
			Body:        responseBody,
		}
		api.cache.Set(targetURL, cachedResp)
	}

	_, _ = w.Write(responseBody) //nolint:gosec // G705: proxy passthrough by design
}

func (api *api) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	cacheSize := api.cache.Size()
	response := fmt.Sprintf(`{"status":"ok","cache_entries":%d}`, cacheSize)
	_, _ = w.Write([]byte(response))
}

func (api *api) handleCacheClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	api.cache.Clear()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := `{"status":"ok","message":"cache cleared"}`
	_, _ = w.Write([]byte(response))
}
