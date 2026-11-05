package caches

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/configs"
)

func decodeCacheValue[T any](value interface{}) (*T, error) {
	if value == nil {
		return nil, fmt.Errorf("empty cache value")
	}
	if typed, ok := value.(*T); ok {
		return typed, nil
	}
	if typed, ok := value.(T); ok {
		return &typed, nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	out := new(T)
	if err = json.Unmarshal(raw, out); err != nil {
		return nil, fmt.Errorf("decode cache value: %w", err)
	}
	return out, nil
}

const defaultCacheTTL = 15 * time.Minute

func getTTL(cacheConfig map[string]configs.CacheConfig, key string) time.Duration {
	cfg, ok := cacheConfig[key]
	if !ok {
		return defaultCacheTTL
	}
	ttl := time.Duration(cfg.TTLDuration) * unitParser(cfg.TTLUnit)
	if ttl <= 0 {
		return defaultCacheTTL
	}
	return ttl
}

func unitParser(unit string) time.Duration {
	switch unit {
	case "seconds":
		return time.Second
	case "minutes":
		return time.Minute
	case "hours":
		return time.Hour
	}
	return time.Duration(0)
}
