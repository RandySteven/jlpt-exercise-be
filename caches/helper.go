package caches

import (
	"encoding/json"
	"fmt"
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
