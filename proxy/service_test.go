package proxy

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

func TestValidateURL(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("invalid scheme", func(t *testing.T) {
		t.Parallel()
		u, _ := url.Parse("ftp://example.com")
		err := validateURL(ctx, u)
		if err == nil {
			t.Error("expected error for invalid scheme")
		}
	})

	t.Run("private IP", func(t *testing.T) {
		t.Parallel()
		u, _ := url.Parse("http://127.0.0.1")
		err := validateURL(ctx, u)
		if err == nil {
			t.Error("expected error for private IP")
		}
	})
}

func TestService_URL(t *testing.T) {
	origRanges := privateIPRanges
	privateIPRanges = []*net.IPNet{}
	defer func() { privateIPRanges = origRanges }()

	service := New()
	ctx := context.Background()

	t.Run("non-m3u8 proxy", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("hello world"))
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		res, err := service.URL(ctx, u)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.ContentType != "text/plain" {
			t.Errorf("expected text/plain, got %v", res.ContentType)
		}
		if res.Body == nil {
			t.Fatal("expected Body to be non-nil")
		}
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		if string(body) != "hello world" {
			t.Errorf("expected 'hello world', got %v", string(body))
		}
	})

	t.Run("m3u8 proxy uncompressed", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("#EXTM3U\n#EXTINF:10.0,\nhttp://example.com/segment.ts"))
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		res, err := service.URL(ctx, u)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Body != nil {
			t.Error("expected Body to be nil for m3u8")
		}
		if string(res.Content) != "#EXTM3U\n#EXTINF:10.0,\nhttp://example.com/segment.ts" {
			t.Errorf("unexpected content: %v", string(res.Content))
		}
	})

	t.Run("m3u8 proxy gzip compressed", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.Header().Set("Content-Encoding", "gzip")
			w.WriteHeader(http.StatusOK)

			var b bytes.Buffer
			gz := gzip.NewWriter(&b)
			gz.Write([]byte("#EXTM3U"))
			gz.Close()
			w.Write(b.Bytes())
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		res, err := service.URL(ctx, u)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(res.Content) != "#EXTM3U" {
			t.Errorf("expected uncompressed content, got %v", string(res.Content))
		}
		if _, ok := res.Headers["Content-Encoding"]; ok {
			t.Error("expected Content-Encoding header to be removed")
		}
	})

	t.Run("m3u8 proxy deflate compressed", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.Header().Set("Content-Encoding", "deflate")
			w.WriteHeader(http.StatusOK)

			var b bytes.Buffer
			fl, _ := flate.NewWriter(&b, flate.DefaultCompression)
			fl.Write([]byte("#EXTM3U"))
			fl.Close()
			w.Write(b.Bytes())
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		res, err := service.URL(ctx, u)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(res.Content) != "#EXTM3U" {
			t.Errorf("expected uncompressed content, got %v", string(res.Content))
		}
	})

	t.Run("m3u8 proxy zlib compressed", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.Header().Set("Content-Encoding", "zlib")
			w.WriteHeader(http.StatusOK)

			var b bytes.Buffer
			zl := zlib.NewWriter(&b)
			zl.Write([]byte("#EXTM3U"))
			zl.Close()
			w.Write(b.Bytes())
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		res, err := service.URL(ctx, u)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(res.Content) != "#EXTM3U" {
			t.Errorf("expected uncompressed content, got %v", string(res.Content))
		}
	})

	t.Run("m3u8 proxy br compressed", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.Header().Set("Content-Encoding", "br")
			w.WriteHeader(http.StatusOK)

			var b bytes.Buffer
			br := brotli.NewWriter(&b)
			br.Write([]byte("#EXTM3U"))
			br.Close()
			w.Write(b.Bytes())
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		res, err := service.URL(ctx, u)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(res.Content) != "#EXTM3U" {
			t.Errorf("expected uncompressed content, got %v", string(res.Content))
		}
	})

	t.Run("m3u8 proxy zstd compressed", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.Header().Set("Content-Encoding", "zstd")
			w.WriteHeader(http.StatusOK)

			var b bytes.Buffer
			z, _ := zstd.NewWriter(&b)
			z.Write([]byte("#EXTM3U"))
			z.Close()
			w.Write(b.Bytes())
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		res, err := service.URL(ctx, u)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(res.Content) != "#EXTM3U" {
			t.Errorf("expected uncompressed content, got %v", string(res.Content))
		}
	})

	t.Run("m3u8 URL suffix but generic content type", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("#EXTM3U"))
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL + "/playlist.m3u8")
		res, err := service.URL(ctx, u)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Body != nil {
			t.Error("expected Body to be nil because URL is recognized as m3u8")
		}
		if string(res.Content) != "#EXTM3U" {
			t.Errorf("unexpected content: %v", string(res.Content))
		}
	})

	t.Run("server returns error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		res, err := service.URL(ctx, u)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Status != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", res.Status)
		}
	})

	t.Run("unreachable server", func(t *testing.T) {
		u, _ := url.Parse("http://127.0.0.1:0") // port 0 should fail to connect
		_, err := service.URL(ctx, u)
		if err == nil {
			t.Error("expected error for unreachable server")
		}
	})
}

func TestService_ProcessM3U8(t *testing.T) {
	t.Parallel()
	service := New()
	proxyURL := "http://proxy.local/proxy"

	tests := []struct {
		name     string
		content  string
		baseURL  string
		expected string
	}{
		{
			name:     "absolute url",
			content:  "#EXTM3U\nhttp://example.com/segment.ts",
			baseURL:  "http://example.com/playlist.m3u8",
			expected: fmt.Sprintf("#EXTM3U\n%s?url=%s", proxyURL, url.QueryEscape("http://example.com/segment.ts")),
		},
		{
			name:     "relative url",
			content:  "#EXTM3U\nsegment.ts",
			baseURL:  "http://example.com/path/playlist.m3u8",
			expected: fmt.Sprintf("#EXTM3U\n%s?url=%s", proxyURL, url.QueryEscape("http://example.com/path/segment.ts")),
		},
		{
			name:     "relative url with trailing slash base",
			content:  "#EXTM3U\nsegment.ts",
			baseURL:  "http://example.com/path/",
			expected: fmt.Sprintf("#EXTM3U\n%s?url=%s", proxyURL, url.QueryEscape("http://example.com/path/segment.ts")),
		},
		{
			name:     "protocol relative url",
			content:  "#EXTM3U\n//example.com/segment.ts",
			baseURL:  "https://other.com/playlist.m3u8",
			expected: fmt.Sprintf("#EXTM3U\n%s?url=%s", proxyURL, url.QueryEscape("https://example.com/segment.ts")),
		},
		{
			name:     "url with URI attribute",
			content:  `#EXT-X-KEY:METHOD=AES-128,URI="http://example.com/key.bin"`,
			baseURL:  "http://example.com/playlist.m3u8",
			expected: fmt.Sprintf(`#EXT-X-KEY:METHOD=AES-128,URI="%s?url=%s"`, proxyURL, url.QueryEscape("http://example.com/key.bin")),
		},
		{
			name:     "relative url with URI attribute",
			content:  `#EXT-X-KEY:METHOD=AES-128,URI="key.bin"`,
			baseURL:  "http://example.com/path/playlist.m3u8",
			expected: fmt.Sprintf(`#EXT-X-KEY:METHOD=AES-128,URI="%s?url=%s"`, proxyURL, url.QueryEscape("http://example.com/path/key.bin")),
		},
		{
			name:     "already proxied URL",
			content:  fmt.Sprintf("#EXTM3U\n%s?url=http%%3A%%2F%%2Fexample.com", proxyURL),
			baseURL:  "http://example.com/playlist.m3u8",
			expected: fmt.Sprintf("#EXTM3U\n%s?url=http%%3A%%2F%%2Fexample.com", proxyURL),
		},
		{
			name:     "already proxied URI attribute",
			content:  fmt.Sprintf(`#EXT-X-KEY:METHOD=AES-128,URI="%s?url=http%%3A%%2F%%2Fexample.com"`, proxyURL),
			baseURL:  "http://example.com/playlist.m3u8",
			expected: fmt.Sprintf(`#EXT-X-KEY:METHOD=AES-128,URI="%s?url=http%%3A%%2F%%2Fexample.com"`, proxyURL),
		},
		{
			name:     "normalize malformed URL",
			content:  "#EXTM3U\nhttps//example.com/segment.ts",
			baseURL:  "http://example.com/playlist.m3u8",
			expected: fmt.Sprintf("#EXTM3U\n%s?url=%s", proxyURL, url.QueryEscape("https://example.com/segment.ts")),
		},
		{
			name:     "normalize malformed URL in URI",
			content:  `#EXT-X-KEY:METHOD=AES-128,URI="https//example.com/key.bin"`,
			baseURL:  "http://example.com/playlist.m3u8",
			expected: fmt.Sprintf(`#EXT-X-KEY:METHOD=AES-128,URI="%s?url=%s"`, proxyURL, url.QueryEscape("https://example.com/key.bin")),
		},
		{
			name:     "base url no trailing slash",
			content:  "#EXTM3U\nsegment.ts",
			baseURL:  "http://example.com/path",
			expected: fmt.Sprintf("#EXTM3U\n%s?url=%s", proxyURL, url.QueryEscape("http://example.com/path/segment.ts")),
		},
		{
			name:     "empty lines",
			content:  "#EXTM3U\n\nsegment.ts\n\n",
			baseURL:  "http://example.com/playlist.m3u8",
			expected: fmt.Sprintf("#EXTM3U\n\n%s?url=%s\n", proxyURL, url.QueryEscape("http://example.com/segment.ts")),
		},
		{
			name:     "protocol relative URI attribute",
			content:  `#EXT-X-KEY:METHOD=AES-128,URI="//example.com/key.bin"`,
			baseURL:  "https://other.com/playlist.m3u8",
			expected: fmt.Sprintf(`#EXT-X-KEY:METHOD=AES-128,URI="%s?url=%s"`, proxyURL, url.QueryEscape("https://example.com/key.bin")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u, _ := url.Parse(tt.baseURL)
			result := service.ProcessM3U8([]byte(tt.content), u, proxyURL)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDecompressContent(t *testing.T) {
	t.Parallel()
	service := New()

	t.Run("empty encoding", func(t *testing.T) {
		t.Parallel()
		res, err := service.DecompressContent([]byte("test"), "")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != "test" {
			t.Errorf("expected test, got %s", res)
		}
	})

	t.Run("unknown encoding", func(t *testing.T) {
		t.Parallel()
		res, err := service.DecompressContent([]byte("test"), "unknown")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != "test" {
			t.Errorf("expected test, got %s", res)
		}
	})

	t.Run("invalid gzip data", func(t *testing.T) {
		t.Parallel()
		_, err := service.DecompressContent([]byte("not gzip"), "gzip")
		if err == nil {
			t.Error("expected error for invalid gzip data")
		}
	})
}

func TestValidationError(t *testing.T) {
	t.Parallel()
	err := &ValidationError{err: errors.New("test error")}
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got %v", err.Error())
	}

	errEmpty := &ValidationError{err: nil}
	if errEmpty.Error() != "" {
		t.Errorf("expected empty string, got %v", errEmpty.Error())
	}
}

func TestIsActualM3U8Content(t *testing.T) {
	t.Parallel()
	if !IsActualM3U8Content([]byte("#EXTM3U\n...")) {
		t.Error("expected true for #EXTM3U")
	}
	if !IsActualM3U8Content([]byte("#EXT-X-STREAM-INF\n...")) {
		t.Error("expected true for #EXT-X-STREAM-INF")
	}
	if IsActualM3U8Content([]byte("hello world")) {
		t.Error("expected false for random text")
	}
}

func TestDecompressContent_InvalidData(t *testing.T) {
	t.Parallel()
	service := New()
	invalidData := []byte("this is not compressed data")

	t.Run("invalid deflate", func(t *testing.T) {
		t.Parallel()
		_, err := service.DecompressContent(invalidData, "deflate")
		if err == nil {
			t.Error("expected error for invalid deflate data")
		}
	})

	t.Run("invalid zlib", func(t *testing.T) {
		t.Parallel()
		_, err := service.DecompressContent(invalidData, "zlib")
		if err == nil {
			t.Error("expected error for invalid zlib data")
		}
	})

	t.Run("invalid zstd", func(t *testing.T) {
		t.Parallel()
		_, err := service.DecompressContent(invalidData, "zstd")
		if err == nil {
			t.Error("expected error for invalid zstd data")
		}
	})
}

func TestValidateURL_LookupFailure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("nonexistent host", func(t *testing.T) {
		t.Parallel()
		u, _ := url.Parse("http://this-host-surely-does-not-exist.local")
		err := validateURL(ctx, u)
		if err == nil {
			t.Error("expected error for nonexistent host")
		}
	})
}

func TestService_URL_ContextCanceled(t *testing.T) {
	t.Parallel()
	service := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	u, _ := url.Parse("http://example.com")
	_, err := service.URL(ctx, u)
	if err == nil {
		t.Error("expected error due to canceled context")
	}
}
