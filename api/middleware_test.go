package api

import (
	"testing"
	"time"
)

func TestHumanDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"microseconds", 532 * time.Microsecond, "532µs"},
		{"milliseconds", 23 * time.Millisecond, "23.00ms"},
		{"seconds", 2300 * time.Millisecond, "2.30s"},
		{"minutes", 72 * time.Second, "1m12.0s"},
		{"1ms_boundary", 1000 * time.Microsecond, "1.00ms"},
		{"1s_boundary", 1 * time.Second, "1.00s"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := humanDuration(tc.d)
			if got != tc.want {
				t.Fatalf("humanDuration(%v) = %q; want %q", tc.d, got, tc.want)
			}
		})
	}
}
