package caches

import "github.com/redis/go-redis/v9"

type Caches struct {
}

func NewCaches(redis *redis.Client) *Caches {
	return &Caches{}
}
