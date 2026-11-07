package proxy

import (
	"net/url"
	"path"
	"strings"
)

// VTTContentType is the content type sent to clients for WebVTT responses.
const VTTContentType = "text/vtt; charset=utf-8"

var vttImageExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
	".gif":  {},
	".avif": {},
}

func IsVTTURL(rawURL string) bool {
	rawURL = strings.ToLower(rawURL)
	return strings.HasSuffix(rawURL, ".vtt") ||
		strings.Contains(rawURL, ".vtt?")
}

func IsVTTContentType(contentType string) bool {
	mediaType, _, _ := strings.Cut(strings.ToLower(contentType), ";")
	return strings.TrimSpace(mediaType) == "text/vtt"
}

func IsActualVTTContent(content []byte) bool {
	contentStr := strings.TrimPrefix(string(content), "\ufeff")
	return strings.HasPrefix(contentStr, "WEBVTT")
}

// ProcessVTT routes resource references in cue payloads, such as thumbnail sprites, through the proxy.
func (s *Service) ProcessVTT(content []byte, baseURL *url.URL, proxyURL string, customOrigin, customReferer string) string {
	querySuffix := proxyQuerySuffix(customOrigin, customReferer)

	lines := strings.Split(string(content), "\n")
	inCue := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			inCue = false
		case strings.Contains(trimmed, "-->"):
			inCue = true
		case inCue:
			if proxied, ok := proxyVTTReference(trimmed, baseURL, proxyURL, querySuffix); ok {
				lines[i] = proxied + line[len(strings.TrimRight(line, "\r")):]
			}
		}
	}

	return strings.Join(lines, "\n")
}

// proxyVTTReference keeps the URL fragment (for example "#xywh=...") outside the proxied URL for players.
func proxyVTTReference(line string, baseURL *url.URL, proxyURL, querySuffix string) (string, bool) {
	if strings.ContainsAny(line, " \t<>") || strings.HasPrefix(line, proxyURL) {
		return "", false
	}

	ref, err := url.Parse(normalizeURL(line))
	if err != nil {
		return "", false
	}

	isAbsolute := ref.Scheme == "http" || ref.Scheme == "https"
	_, isImage := vttImageExtensions[strings.ToLower(path.Ext(ref.Path))]
	if ref.ForceQuery || (!isAbsolute && (ref.Scheme != "" || !isImage)) {
		return "", false
	}

	fragment := ref.EscapedFragment()
	ref.Fragment = ""
	ref.RawFragment = ""

	proxied := proxyURL + "?url=" + url.QueryEscape(baseURL.ResolveReference(ref).String()) + querySuffix
	if fragment != "" {
		proxied += "#" + fragment
	}

	return proxied, true
}

func proxyQuerySuffix(customOrigin, customReferer string) string {
	var querySuffix string
	if customOrigin != "" {
		querySuffix += "&origin=" + url.QueryEscape(customOrigin)
	}
	if customReferer != "" {
		querySuffix += "&referer=" + url.QueryEscape(customReferer)
	}

	return querySuffix
}
