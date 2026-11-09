package caches

import (
	redis_client "github.com/RandySteven/go-cook/redis"
	"github.com/RandySteven/jlpt-exercise-be/configs"
)

type Caches struct {
	OnboardingCache       OnboardingCache
	ExerciseCache         ExerciseCache
	SubscriptionCache     SubscriptionCache
	SubscriptionPlanCache SubscriptionPlanCache
}

func NewCaches(redis redis_client.Redis, cacheConfig map[string]configs.CacheConfig) *Caches {
	return &Caches{
		OnboardingCache:       newOnboardingCache(redis, cacheConfig),
		ExerciseCache:         newExerciseCache(redis, cacheConfig),
		SubscriptionCache:     newSubscriptionCache(redis, cacheConfig),
		SubscriptionPlanCache: newSubscriptionPlanCache(redis, cacheConfig),
	}
}
