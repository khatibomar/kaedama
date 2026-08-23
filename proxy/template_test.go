package proxy

import (
	"net/url"
	"reflect"
	"regexp"
	"testing"
)

func TestFindDomainTemplate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		hostname string
		wantFind bool
	}{
		{
			name:     "match exact",
			hostname: "premilkyway.com",
			wantFind: false, // The regex requires a leading dot.
		},
		{
			name:     "match sub domain",
			hostname: "a.premilkyway.com",
			wantFind: true,
		},
		{
			name:     "no match",
			hostname: "example.com",
			wantFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FindDomainTemplate(tt.hostname)
			if (got != nil) != tt.wantFind {
				t.Errorf("FindDomainTemplate() = %v, wantFind %v", got, tt.wantFind)
			}
		})
	}
}

func TestGenerateHeaders(t *testing.T) {
	// Backup original templates and restore after tests
	originalTemplates := templates
	defer func() { templates = originalTemplates }()

	// Add test templates
	templates = append(templates, template{
		pattern:           regexp.MustCompile(`(?i)custom-ua\.com$`),
		origin:            "https://custom.com",
		referer:           "https://custom.com/",
		userAgent:         "CustomUserAgent/1.0",
		additionalHeaders: map[string]string{"x-custom": "value"},
	})

	templates = append(templates, template{
		pattern:           regexp.MustCompile(`(?i)standard\.com$`),
		origin:            "https://standard.com",
		referer:           "https://standard.com/",
		userAgent:         "",
		additionalHeaders: nil,
	})

	tests := []struct {
		name        string
		urlStr      string
		wantHeaders map[string]string
	}{
		{
			name:   "no template matched",
			urlStr: "https://example.com/path",
			wantHeaders: map[string]string{
				headerUserAgent:   defaultUserAgent,
				"accept":          "*/*",
				"accept-language": "en-US,en;q=0.5",
				"sec-fetch-dest":  "empty",
				"sec-fetch-mode":  "cors",
				"sec-fetch-site":  "cross-site",
				headerHost:        "example.com",
			},
		},
		{
			name:   "standard template matched",
			urlStr: "https://standard.com/path",
			wantHeaders: map[string]string{
				headerUserAgent:   defaultUserAgent,
				"accept":          "*/*",
				"accept-language": "en-US,en;q=0.5",
				"sec-fetch-dest":  "empty",
				"sec-fetch-mode":  "cors",
				"sec-fetch-site":  "cross-site",
				headerHost:        "standard.com",
				"origin":          "https://standard.com",
				"referer":         "https://standard.com/",
			},
		},
		{
			name:   "template with custom user agent and additional headers",
			urlStr: "https://custom-ua.com/path",
			wantHeaders: map[string]string{
				headerUserAgent:   "CustomUserAgent/1.0",
				"accept":          "*/*",
				"accept-language": "en-US,en;q=0.5",
				"sec-fetch-dest":  "empty",
				"sec-fetch-mode":  "cors",
				"sec-fetch-site":  "cross-site",
				headerHost:        "custom-ua.com",
				"origin":          "https://custom.com",
				"referer":         "https://custom.com/",
				"x-custom":        "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// No t.Parallel() here because it depends on the mocked global templates.
			u, err := url.Parse(tt.urlStr)
			if err != nil {
				t.Fatalf("Failed to parse URL: %v", err)
			}

			got := GenerateHeaders(u)

			if !reflect.DeepEqual(got, tt.wantHeaders) {
				t.Errorf("GenerateHeaders() = \n%v\nwant \n%v", got, tt.wantHeaders)
			}
		})
	}
}
