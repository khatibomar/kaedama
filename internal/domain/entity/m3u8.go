package entity

import "net/url"

// TODO subject to big changes
type M3U8Entry struct {
	Content  string
	URL      url.URL
	ProxyURL url.URL
}
