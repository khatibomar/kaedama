package proxy

import (
	"net/url"
	"testing"
)

func TestService_ProcessVTT(t *testing.T) {
	t.Parallel()

	base, _ := url.Parse("https://subs.example.com/abc/preview.vtt")
	service := New()

	tests := []struct {
		name    string
		input   string
		origin  string
		referer string
		want    string
	}{
		{
			name:  "relative sprite keeps fragment",
			input: "WEBVTT\n\n00:00:00.000 --> 00:00:05.000\npreview-01.jpg#xywh=0,0,320,180\n",
			want: "WEBVTT\n\n00:00:00.000 --> 00:00:05.000\n" +
				"/proxy?url=https%3A%2F%2Fsubs.example.com%2Fabc%2Fpreview-01.jpg#xywh=0,0,320,180\n",
		},
		{
			name: "absolute and protocol-relative references",
			input: "WEBVTT\n\n1\n00:00.000 --> 00:05.000\nhttps://cdn.example.com/a.jpg\n\n" +
				"00:05.000 --> 00:10.000\n//cdn.example.com/b.png\n",
			want: "WEBVTT\n\n1\n00:00.000 --> 00:05.000\n/proxy?url=https%3A%2F%2Fcdn.example.com%2Fa.jpg\n\n" +
				"00:05.000 --> 00:10.000\n/proxy?url=https%3A%2F%2Fcdn.example.com%2Fb.png\n",
		},
		{
			name:  "subtitle text unchanged",
			input: "WEBVTT\n\n00:00.000 --> 00:05.000\nHello there.jpg\n<i>image.png</i>\nwhat.jpg?\nnote\n",
			want:  "WEBVTT\n\n00:00.000 --> 00:05.000\nHello there.jpg\n<i>image.png</i>\nwhat.jpg?\nnote\n",
		},
		{
			name:  "header lines are not rewritten",
			input: "WEBVTT\nlogo.png\n\nNOTE x.png\n",
			want:  "WEBVTT\nlogo.png\n\nNOTE x.png\n",
		},
		{
			name:    "origin and referer are forwarded",
			input:   "WEBVTT\n\n00:00.000 --> 00:05.000\nx.webp\n",
			origin:  "https://o.example",
			referer: "https://r.example/",
			want: "WEBVTT\n\n00:00.000 --> 00:05.000\n/proxy?url=https%3A%2F%2Fsubs.example.com%2Fabc%2Fx.webp" +
				"&origin=https%3A%2F%2Fo.example&referer=https%3A%2F%2Fr.example%2F\n",
		},
		{
			name:  "crlf line endings are kept",
			input: "WEBVTT\r\n\r\n00:00.000 --> 00:05.000\r\nx.jpg\r\n",
			want:  "WEBVTT\r\n\r\n00:00.000 --> 00:05.000\r\n/proxy?url=https%3A%2F%2Fsubs.example.com%2Fabc%2Fx.jpg\r\n",
		},
		{
			name:  "already proxied reference unchanged",
			input: "WEBVTT\n\n00:00.000 --> 00:05.000\n/proxy?url=x.jpg\n",
			want:  "WEBVTT\n\n00:00.000 --> 00:05.000\n/proxy?url=x.jpg\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := service.ProcessVTT([]byte(tt.input), base, "/proxy", tt.origin, tt.referer)
			if got != tt.want {
				t.Errorf("got\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

func TestVTTDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{"vtt url", IsVTTURL("https://a.com/x.VTT"), true},
		{"vtt url with query", IsVTTURL("https://a.com/x.vtt?t=1"), true},
		{"non vtt url", IsVTTURL("https://a.com/x.vttx"), false},
		{"vtt content type", IsVTTContentType("text/vtt; charset=utf-8"), true},
		{"other content type", IsVTTContentType("text/plain"), false},
		{"vtt content", IsActualVTTContent([]byte("WEBVTT\n")), true},
		{"vtt content with bom", IsActualVTTContent([]byte("\xef\xbb\xbfWEBVTT\n")), true},
		{"html content", IsActualVTTContent([]byte("<!DOCTYPE html>")), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}
