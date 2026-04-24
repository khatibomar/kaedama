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
	"net"
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
		//nolint:gosec // Disable SSL certificate verification
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // This ignores SSL certificate errors
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

var privateIPRanges = buildPrivateIPRanges()

func buildPrivateIPRanges() []*net.IPNet {
	cidrs := []string{
		"0.0.0.0/8",      // "This" network
		"10.0.0.0/8",     // RFC 1918 private
		"100.64.0.0/10",  // CGNAT (RFC 6598)
		"127.0.0.0/8",    // Loopback
		"169.254.0.0/16", // Link-local (incl. AWS/GCP/Azure metadata)
		"172.16.0.0/12",  // RFC 1918 private
		"192.168.0.0/16", // RFC 1918 private
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 unique local
		"fe80::/10",      // IPv6 link-local
	}
	ranges := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Sprintf("invalid CIDR %s: %v", cidr, err))
		}
		ranges = append(ranges, block)
	}
	return ranges
}

func isPrivateIP(ip net.IP) bool {
	for _, block := range privateIPRanges {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func validateURL(ctx context.Context, u *url.URL) error {
	if u.Scheme != "http" && u.Scheme != "https" {
		return &ValidationError{err: fmt.Errorf("invalid scheme %q: only http and https are allowed", u.Scheme)}
	}

	host := u.Hostname()
	addrs, err := net.DefaultResolver.LookupHost(ctx, host)
	if err != nil {
		return &ValidationError{err: fmt.Errorf("failed to resolve host: %w", err)}
	}

	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil {
			continue
		}
		if isPrivateIP(ip) {
			return &ValidationError{err: errors.New("target URL resolves to a private or reserved address")}
		}
	}

	return nil
}

func (s *Service) URL(ctx context.Context, requestURL *url.URL) (*Result, error) {
	if err := validateURL(ctx, requestURL); err != nil {
		return nil, err
	}

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

// normalizeURL fixes malformed protocol-relative-looking URLs like
// "https//host/path" → "https://host/path"
func normalizeURL(rawURL string) string {
	for _, scheme := range []string{"https", "http"} {
		malformed := scheme + "//"
		if strings.HasPrefix(rawURL, malformed) {
			return scheme + "://" + rawURL[len(malformed):]
		}
	}
	return rawURL
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

			if len(matches) == 0 {
				// Comment line with no URI tag — preserve as-is
				processedLines = append(processedLines, line)
				continue
			}

			for _, match := range matches {
				rawURI := match[1]
				originalURI := normalizeURL(rawURI)

				if strings.HasPrefix(originalURI, proxyURL) {
					continue
				}

				var absoluteURI string
				if strings.HasPrefix(originalURI, "http://") || strings.HasPrefix(originalURI, "https://") {
					absoluteURI = originalURI
				} else if strings.HasPrefix(originalURI, "//") {
					absoluteURI = fmt.Sprintf("%s:%s", baseURL.Scheme, originalURI)
				} else {
					absoluteURI = fmt.Sprintf("%s%s", basePath, originalURI)
				}

				proxiedURL := fmt.Sprintf("%s?url=%s", proxyURL, url.QueryEscape(absoluteURI))
				line = strings.ReplaceAll(line,
					fmt.Sprintf(`URI="%s"`, rawURI),
					fmt.Sprintf(`URI="%s"`, proxiedURL),
				)
			}
			processedLines = append(processedLines, line)
			continue
		}

		if len(trimmedLine) > 0 {
			trimmedLine = normalizeURL(trimmedLine)

			if strings.HasPrefix(trimmedLine, proxyURL) {
				processedLines = append(processedLines, line)
				continue
			}

			var absoluteURL string
			if strings.HasPrefix(trimmedLine, "http://") || strings.HasPrefix(trimmedLine, "https://") {
				absoluteURL = trimmedLine
			} else if strings.HasPrefix(trimmedLine, "//") {
				absoluteURL = fmt.Sprintf("%s:%s", baseURL.Scheme, trimmedLine)
			} else {
				absoluteURL = fmt.Sprintf("%s%s", basePath, trimmedLine)
			}

			proxiedURL := fmt.Sprintf("%s?url=%s", proxyURL, url.QueryEscape(absoluteURL))
			processedLines = append(processedLines, proxiedURL)
			continue
		}

		// Empty line — preserve it
		processedLines = append(processedLines, line)
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

func IsActualM3U8Content(content []byte) bool {
	contentStr := strings.TrimSpace(string(content))
	res := strings.HasPrefix(contentStr, "#EXTM3U") || strings.HasPrefix(contentStr, "#EXT-X-STREAM-INF")
	return res
}
