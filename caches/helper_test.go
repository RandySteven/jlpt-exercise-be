package caches

import (
	"testing"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/configs"
)

func TestGetTTL(t *testing.T) {
	config := map[string]configs.CacheConfig{
		"exercise": {TTLDuration: 600, TTLUnit: "seconds"},
		"plans":    {TTLDuration: 1, TTLUnit: "hours"},
		"blank":    {TTLDuration: 0, TTLUnit: "seconds"},
		"typo":     {TTLDuration: 30, TTLUnit: "second"},
		"negative": {TTLDuration: -5, TTLUnit: "minutes"},
	}

	tests := []struct {
		key  string
		want time.Duration
	}{
		{key: "exercise", want: 600 * time.Second},
		{key: "plans", want: time.Hour},
		{key: "missing", want: defaultCacheTTL},
		{key: "blank", want: defaultCacheTTL},
		{key: "typo", want: defaultCacheTTL},
		{key: "negative", want: defaultCacheTTL},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := getTTL(config, tt.key); got != tt.want {
				t.Fatalf("getTTL(%q) = %s, want %s", tt.key, got, tt.want)
			}
		})
	}

	if got := getTTL(nil, "exercise"); got != defaultCacheTTL {
		t.Fatalf("nil config TTL = %s, want %s", got, defaultCacheTTL)
	}
}
