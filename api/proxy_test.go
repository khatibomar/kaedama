package api

import "testing"

func TestBuildCacheKey(t *testing.T) {
	t.Parallel()

	const target = "https://example.com/stream.m3u8"

	t.Run("no overrides keeps the bare target", func(t *testing.T) {
		t.Parallel()

		if got := buildCacheKey(target, "", ""); got != target {
			t.Errorf("expected %q, got %q", target, got)
		}
	})

	t.Run("overrides scope the key", func(t *testing.T) {
		t.Parallel()

		base := buildCacheKey(target, "", "")
		withOrigin := buildCacheKey(target, "https://a.example", "")
		withReferer := buildCacheKey(target, "", "https://b.example/")
		withBoth := buildCacheKey(target, "https://a.example", "https://b.example/")

		keys := map[string]struct{}{
			base:        {},
			withOrigin:  {},
			withReferer: {},
			withBoth:    {},
		}
		if len(keys) != 4 {
			t.Errorf("expected 4 distinct cache keys, got %d", len(keys))
		}
	})

	t.Run("same overrides produce the same key", func(t *testing.T) {
		t.Parallel()

		a := buildCacheKey(target, "https://a.example", "https://b.example/")
		b := buildCacheKey(target, "https://a.example", "https://b.example/")
		if a != b {
			t.Errorf("expected %q, got %q", a, b)
		}
	})
}
