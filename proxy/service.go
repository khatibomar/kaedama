package proxy

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

type Service struct {
	client *http.Client
}

func New() *Service {
	// Create HTTP client with custom transport for better performance
	transport := &http.Transport{
		//nolint:gosec // I don't know sec stuff.
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 30,
		IdleConnTimeout:     90 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &Service{
		client: client,
	}
}

type ValidationError struct {
	err error
}

func (e ValidationError) Error() string {
	if e.err == nil {
		return ""
	}

	return e.err.Error()
}

type Result struct {
	ContentType string
	Headers     map[string]string
	Status      int
	Content     []byte
}

func (s *Service) URL(ctx context.Context, requestURL *url.URL) (*Result, error) {
	targetURL := requestURL.String()

	headers := GenerateHeaders(requestURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := s.client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, errors.New("request timed out after 30s")
		}
		return nil, fmt.Errorf("proxy request failed: %w", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	encoding := resp.Header.Get("Content-Encoding")
	if encoding != "" {
		decompressed, err := s.DecompressContent(content, encoding)
		if err == nil {
			content = decompressed
		}
	}

	headers = make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return &Result{
		ContentType: contentType,
		Headers:     headers,
		Status:      resp.StatusCode,
		Content:     content,
	}, nil
}

func (s *Service) ProcessM3U8(content []byte, baseURL *url.URL, proxyURL string) string {
	basePath := baseURL.String()

	if strings.HasSuffix(basePath, ".m3u8") {
		basePath = basePath[:strings.LastIndex(basePath, "/")+1]
	} else if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	lines := strings.Split(string(content), "\n")
	processedLines := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "#") {
			re := regexp.MustCompile(`URI="([^"]+)"`)
			matches := re.FindAllStringSubmatch(line, -1)

			for _, match := range matches {
				originalURL := match[1]
				if strings.HasPrefix(originalURL, proxyURL) {
					continue
				}

				var absoluteURL string
				if strings.HasPrefix(originalURL, "http://") || strings.HasPrefix(originalURL, "https://") {
					absoluteURL = originalURL
				} else if strings.HasPrefix(originalURL, "//") {
					// Protocol-relative URL
					absoluteURL = fmt.Sprintf("%s%s", baseURL.Scheme, originalURL)
				} else {
					// Relative URL
					absoluteURL = fmt.Sprintf("%s%s", basePath, originalURL)
				}

				proxiedURL := fmt.Sprintf("%s?url=%s", proxyURL, url.QueryEscape(absoluteURL))

				line = strings.ReplaceAll(line, fmt.Sprintf(`URI="%s"`, originalURL), fmt.Sprintf(`URI="%s"`, proxiedURL))
				processedLines = append(processedLines, line)
				continue
			}
			processedLines = append(processedLines, line)
		}

		if len(trimmedLine) > 0 {
			if strings.HasPrefix(trimmedLine, proxyURL) {
				processedLines = append(processedLines, line)
				continue
			}

			var absoluteURL string
			if strings.HasPrefix(trimmedLine, "http://") || strings.HasPrefix(trimmedLine, "https://") {
				absoluteURL = trimmedLine
			} else if strings.HasPrefix(trimmedLine, "//") {
				// Protocol-relative URL
				absoluteURL = fmt.Sprintf("%s%s", baseURL.Scheme, trimmedLine)
			} else {
				// Relative URL
				absoluteURL = fmt.Sprintf("%s%s", basePath, trimmedLine)
			}
			// Create proxy URL
			proxiedURL := fmt.Sprintf("%s?url=%s", proxyURL, url.QueryEscape(absoluteURL))

			processedLines = append(processedLines, proxiedURL)
		}
	}

	result := strings.Join(processedLines, "\n")

	if unescaped, err := url.QueryUnescape(result); err == nil {
		return unescaped
	}

	return result
}

// DecompressContent decompresses content if needed (gzip, deflate, zlib,
// brotli, zstd).
func (s *Service) DecompressContent(content []byte, encoding string) ([]byte, error) {
	if encoding == "" {
		return content, nil
	}

	switch strings.ToLower(encoding) {
	case "gzip":
		reader, err := gzip.NewReader(bytes.NewReader(content))
		if err != nil {
			return content, err
		}
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return content, err
		}
		return decompressed, nil

	case "deflate":
		reader := flate.NewReader(bytes.NewReader(content))
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return content, err
		}
		return decompressed, nil

	case "zlib":
		reader, err := zlib.NewReader(bytes.NewReader(content))
		if err != nil {
			return content, err
		}
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return content, err
		}
		return decompressed, nil

	case "br":
		reader := brotli.NewReader(bytes.NewReader(content))

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return content, err
		}
		return decompressed, nil

	case "zstd", "zst":
		decoder, err := zstd.NewReader(bytes.NewReader(content))
		if err != nil {
			return content, err
		}
		defer decoder.Close()

		decompressed, err := io.ReadAll(decoder)
		if err != nil {
			return content, err
		}
		return decompressed, nil
	default:
		return content, nil
	}
}

func IsM3U8URL(url string) bool {
	url = strings.ToLower(url)
	return strings.HasSuffix(url, ".m3u8") ||
		strings.Contains(url, ".m3u8?")
}

func IsM3U8ContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	switch contentType {
	case "application/vnd.apple.mpegurl",
		"application/x-mpegurl",
		"audio/mpegurl",
		"audio/x-mpegurl":
		return true
	}
	return false
}
