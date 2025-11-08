package proxy

import (
	"maps"
	"net/url"
	"regexp"
	"strings"
)

// template represents a specific header configuration for a domain.
type template struct {
	pattern           *regexp.Regexp
	origin            string
	referer           string
	userAgent         string
	additionalHeaders map[string]string
}

// defaultUserAgent used when a template doesn't specify one.
const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:137.0) Gecko/20100101 Firefox/137.0"

// templates contains domain-specific anti-hotlinking rules.
var templates = []template{
	{regexp.MustCompile(`\.padorupado\.ru$`), "https://kwik.si", "https://kwik.si/", "", nil},
	{regexp.MustCompile(`krussdomi\.com$`), "https://krussdomi.com", "https://hls.krussdomi.com/", "", nil},
	{regexp.MustCompile(`\.narutokun\.xyz$`), "https://krussdomi.com", "https://krussdomi.com/", "", nil},
	{regexp.MustCompile(`\.babybayw\.xyz$`), "https://krussdomi.com", "https://krussdomi.com/", "", nil},
	{regexp.MustCompile(`\.advancedairesearchlab\.xyz$`), "https://krussdomi.com", "https://krussdomi.com/", "", nil},
	{regexp.MustCompile(`\.habibikun\.xyz$`), "https://bl.krussdomi.com", "https://bl.krussdomi.com/", "", nil},
	{regexp.MustCompile(`\.akamaized\.net$`), "https://bl.krussdomi.com", "https://bl.krussdomi.com/", "", nil},
	{regexp.MustCompile(`\.anih1\.top$`), "https://ee.anih1.top", "https://ee.anih1.top/", "", nil},
	{regexp.MustCompile(`\.xyk3\.top$`), "https://ee.anih1.top", "https://ee.anih1.top/", "", nil},
	{regexp.MustCompile(`\.premilkyway\.com$`), "https://uqloads.xyz", "https://uqloads.xyz/", "", nil},
	{regexp.MustCompile(`\.kwikie\.ru$`), "https://kwik.si", "https://kwik.si/", "", nil},
	{
		regexp.MustCompile(`(revolutionizingtheweb|nextgentechnologytrends|smartinvestmentstrategies|` +
			`creativedesignstudioxyz|breakingdigitalboundaries|ultimatetechinnovation)\.xyz$`),
		"https://hls.krussdomi.com", "https://hls.krussdomi.com/", "", nil,
	},
	{regexp.MustCompile(`\.raffaellocdn\.net$`), "https://streameeeeee.site", "https://streameeeeee.site/", "", nil},
	{regexp.MustCompile(`(dewbreeze84|mistyvalley31)\.(online|live)$`), "https://megacloud.blog", "https://megacloud.blog/", "", nil},
	{
		regexp.MustCompile(`douvid\.xyz$`),
		"https://megacloud.blog",
		"https://megacloud.blog/",
		"",
		map[string]string{
			"accept":          "*/*",
			"accept-language": "en-US,en;q=0.5",
			"sec-fetch-dest":  "empty",
			"sec-fetch-mode":  "cors",
			"sec-fetch-site":  "cross-site",
		},
	},
	// (You can continue adding the remaining patterns the same way...)
}

// FindDomainTemplate returns the first template that matches a given hostname.
func FindDomainTemplate(hostname string) *template {
	for _, template := range templates {
		if template.pattern.MatchString(hostname) {
			return &template
		}
	}
	return nil
}

// GenerateHeaders builds a header map for a given URL, using the domain templates.
func GenerateHeaders(u *url.URL) map[string]string {
	headers := map[string]string{
		"user-agent":      defaultUserAgent,
		"accept":          "*/*",
		"accept-language": "en-US,en;q=0.5",
		"sec-fetch-dest":  "empty",
		"sec-fetch-mode":  "cors",
		"sec-fetch-site":  "cross-site",
		"host":            u.Host,
	}

	hostname := strings.ToLower(u.Hostname())
	template := FindDomainTemplate(hostname)
	if template == nil {
		// No template found, use default headers
		return headers
	}

	headers["origin"] = template.origin
	headers["referer"] = template.referer

	if template.userAgent != "" {
		headers["user-agent"] = template.userAgent
	}

	maps.Copy(headers, template.additionalHeaders)

	return headers
}
