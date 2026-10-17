package caches

import redis_client "github.com/RandySteven/go-cook/redis"

type Caches struct {
	OnboardingCache OnboardingCache
}

func NewCaches(redis redis_client.Redis) *Caches {
	return &Caches{
		OnboardingCache: newOnboardingCache(redis),
	}
}
