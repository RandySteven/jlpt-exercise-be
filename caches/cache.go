package caches

import (
	redis_client "github.com/RandySteven/go-cook/redis"
	"github.com/RandySteven/jlpt-exercise-be/configs"
)

type Caches struct {
	OnboardingCache       OnboardingCache
	ExerciseCache         ExerciseCache
	PaymentCache          PaymentCache
	SubscriptionPlanCache SubscriptionPlanCache
}

func NewCaches(redis redis_client.Redis, cacheConfig map[string]configs.CacheConfig) *Caches {
	return &Caches{
		OnboardingCache:       newOnboardingCache(redis, cacheConfig),
		ExerciseCache:         newExerciseCache(redis, cacheConfig),
		PaymentCache:          newPaymentCache(redis, cacheConfig),
		SubscriptionPlanCache: newSubscriptionPlanCache(redis, cacheConfig),
	}
}
