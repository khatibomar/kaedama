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
	// Padorupado.ru
	{regexp.MustCompile(`(?i)\.padorupado\.ru$`), "https://kwik.si", "https://kwik.si/", "", nil},

	// Krussdomi.com related
	{regexp.MustCompile(`(?i)krussdomi\.com$`), "https://krussdomi.com", "https://hls.krussdomi.com/", "", nil},
	{regexp.MustCompile(`(?i)\.narutokun\.xyz$`), "https://krussdomi.com", "https://krussdomi.com/", "", nil},
	{regexp.MustCompile(`(?i)\.babybayw\.xyz$`), "https://krussdomi.com", "https://krussdomi.com/", "", nil},
	{regexp.MustCompile(`(?i)\.advancedairesearchlab\.xyz$`), "https://krussdomi.com", "https://krussdomi.com/", "", nil},
	{regexp.MustCompile(`(?i)\.habibikun\.xyz$`), "https://bl.krussdomi.com", "https://bl.krussdomi.com/", "", nil},
	{regexp.MustCompile(`(?i)\.akamaized\.net$`), "https://bl.krussdomi.com", "https://bl.krussdomi.com/", "", nil},

	// Anih1 related
	{regexp.MustCompile(`(?i)\.anih1\.top$`), "https://ee.anih1.top", "https://ee.anih1.top/", "", nil},
	{regexp.MustCompile(`(?i)\.xyk3\.top$`), "https://ee.anih1.top", "https://ee.anih1.top/", "", nil},

	// Premilkyway
	{regexp.MustCompile(`(?i)\.premilkyway\.com$`), "https://uqloads.xyz", "https://uqloads.xyz/", "", nil},

	// Kwikie.ru
	{regexp.MustCompile(`(?i)\.kwikie\.ru$`), "https://kwik.si", "https://kwik.si/", "", nil},

	// Various xyz domains (krussdomi.com related)
	{
		regexp.MustCompile(`(?i)(revolutionizingtheweb|nextgentechnologytrends|smartinvestmentstrategies|` +
			`creativedesignstudioxyz|breakingdigitalboundaries|ultimatetechinnovation)\.xyz$`),
		"https://hls.krussdomi.com", "https://hls.krussdomi.com/", "", nil,
	},

	// Raffaellocdn.net
	{regexp.MustCompile(`(?i)\.raffaellocdn\.net$`), "https://streameeeeee.site", "https://streameeeeee.site/", "", nil},

	// Megacloud related domains
	{
		regexp.MustCompile(
			`(?i)(dewbreeze84|mistyvalley31)\.(online|live)$`),
		"https://megacloud.blog", "https://megacloud.blog/", "", nil,
	},
	{
		regexp.MustCompile(`(?i)douvid\.xyz$`),
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

	// Lightning/Storm/Weather themed domains (Megacloud related)
	{
		regexp.MustCompile(`(?i)(lightningspark77|thunderwave48|stormwatch95|windyrays29|thunderstrike77|fogtwist21|` +
			`rainfallpath36|lightningflash39|stormwhirl73|cloudburst82|drizzleshower19)\.` +
			`(pro|site|xyz|online|live)$`),
		"https://megacloud.club", "https://megacloud.club/", "", nil,
	},

	// megaplay related
	{
		regexp.MustCompile(`(?i)(mewstream|flareon|zaplume|lumiflow|ovexa|sparqle|voltara|` +
			`flarestorm|zaptrix|lookaround|renvix|gleamwave|zapora|glimmeron)\.` +
			`(click|buzz|live|club)$`),
		"https://megaplay.buzz", "https://megaplay.buzz/", "", nil,
	},

	// Clear sky drift
	{regexp.MustCompile(`(?i)clearskydrift45\.site$`), "https://kerolaunochan.online", "https://kerolaunochan.online/", "", nil},

	// Cloudnestra related
	{regexp.MustCompile(`(?i)\.shadowlandschronicles\.com$`), "https://cloudnestra.com", "https://cloudnestra.com/", "", nil},
	{
		regexp.MustCompile(`(?i)(sparkrisestudios|dreamwavecollective|urbansagecollective|novaquestdynamics|boldsageventures)\.xyz$`),
		"https://cloudnestra.com", "https://cloudnestra.com/", "", nil,
	},
	{regexp.MustCompile(`(?i)putgate\.org$`), "https://cloudnestra.com", "https://cloudnestra.com/", "", nil},

	// Southboat.site
	{regexp.MustCompile(`(?i)\.southboat\.site$`), "https://player.videasy.net", "https://player.videasy.net/", "", nil},

	// Cdnup.cc and streamupcdn.com
	{regexp.MustCompile(`(?i)\.cdnup\.cc$`), "https://bestwish.lol", "https://bestwish.lol/", "", nil},
	{regexp.MustCompile(`(?i)\.streamupcdn\.com$`), "https://bestwish.lol", "https://bestwish.lol/", "", nil},

	// Netmagcdn.com
	{regexp.MustCompile(`(?i)\.netmagcdn\.com$`), "https://megacloud.club", "https://megacloud.club/", "", nil},

	// Vmeas.cloud
	{regexp.MustCompile(`(?i)vmeas\.cloud$`), "https://vidmoly.to", "https://vidmoly.to/", "", nil},

	// Nextwaveinitiative and shadowlandschronicles (edgedeliverynetwork)
	{
		regexp.MustCompile(`(?i)nextwaveinitiative\.xyz$`),
		"https://edgedeliverynetwork.org", "https://edgedeliverynetwork.org/", "", nil,
	},
	{
		regexp.MustCompile(`(?i)shadowlandschronicles\.com$`),
		"https://edgedeliverynetwork.org", "https://edgedeliverynetwork.org/", "", nil,
	},

	// Lightning bolts and vidsrc related
	{regexp.MustCompile(`(?i)lightningbolts\.ru$`), "https://vidsrc.cc", "https://vidsrc.cc/", "", nil},
	{regexp.MustCompile(`(?i)\.xelvonwave64\.xyz$`), "https://vidsrc.su", "https://vidsrc.su/", "", nil},
	{regexp.MustCompile(`(?i)lightningbolt\.site$`), "https://vidsrc.cc", "https://vidsrc.cc/", "", nil},
	{regexp.MustCompile(`(?i)vyebzzqlojvrl\.top$`), "https://vidsrc.cc", "https://vidsrc.cc/", "", nil},

	// Vidlvod.store
	{regexp.MustCompile(`(?i)vidlvod\.store$`), "https://vidlink.pro", "https://vidlink.pro/", "", nil},

	// Megacloud Store domains (extended weather-themed list)
	{
		regexp.MustCompile(`(?i)(sunnybreeze16|crimsonstorm18|glacierfalcon72|mgstatics|cloudydrift38|stormwhirl73|odyssey|rainveil36|` +
			`sunshinerays93|sunburst66|sunburst93|windytrail24|stormshade84|clearskyline88|clearbluesky72|` +
			`breezygale56|haildrop77|frostshine12|frostbite27|frostywinds57|icyhailstorm64|icyhailstorm29|` +
			`windflash93|stormdrift27|tempestcloud61|rainfallpath36)\.` +
			`(live|site|xyz|online|pro|biz|wiki)$`),
		"https://megacloud.blog", "https://megacloud.blog/", "", nil,
	},

	// Odyssey domains (megacloud related)
	{regexp.MustCompile(`(?i)odyssey-\d+\.biz$`), "https://megaup.live", "https://megaup.live/", "", nil},

	// 1stkmgv1.com
	{regexp.MustCompile(`(?i)1stkmgv1\.com$`), "https://vidmoly.to", "https://vidmoly.to/", "", nil},

	// Rainstorm92.xyz
	{regexp.MustCompile(`(?i)rainstorm92\.xyz$`), "https://megacloud.club", "https://megacloud.club/", "", nil},

	// Feetcdn.com
	{regexp.MustCompile(`(?i)\.feetcdn\.com$`), "https://kerolaunochan.online", "https://kerolaunochan.online/", "", nil},

	// Kerolaunochan.live domains
	{
		regexp.MustCompile(`(?i)(heatwave90|humidmist27|frozenbreeze65|drizzlerain73|sunrays81)\.` +
			`(pro|wiki|live|online|xyz)$`),
		"https://kerolaunochan.live", "https://kerolaunochan.live/", "", nil,
	},

	// Embed.su related
	{regexp.MustCompile(`(?i)embed\.su$`), "https://embed.su", "https://embed.su/", "", nil},
	{regexp.MustCompile(`(?i)usbigcdn\.cc$`), "https://embed.su", "https://embed.su/", "", nil},
	{regexp.MustCompile(`(?i)\.congacdn\.cc$`), "https://embed.su", "https://embed.su/", "", nil},

	// Vkcdn5.com
	{regexp.MustCompile(`(?i)\.vkcdn5\.com$`), "https://vkspeed.com", "https://vkspeed.com/", "", nil},

	// Cloudfront CDN
	{
		regexp.MustCompile(`(?i)\.cloudfront\.net$`),
		"https://d2zihajmogu5jn.cloudfront.net", "https://d2zihajmogu5jn.cloudfront.net/", "", nil,
	},

	// Twitch CDN
	{regexp.MustCompile(`(?i)\.ttvnw\.net$`), "https://www.twitch.tv", "https://www.twitch.tv/", "", nil},
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
