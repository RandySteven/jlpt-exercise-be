package caches

import redis_client "github.com/RandySteven/go-cook/redis"

type Caches struct {
	OnboardingCache       OnboardingCache
	ExerciseCache         ExerciseCache
	PaymentCache          PaymentCache
	SubscriptionPlanCache SubscriptionPlanCache
}

func NewCaches(redis redis_client.Redis) *Caches {
	return &Caches{
		OnboardingCache:       newOnboardingCache(redis),
		ExerciseCache:         newExerciseCache(redis),
		PaymentCache:          newPaymentCache(redis),
		SubscriptionPlanCache: newSubscriptionPlanCache(redis),
	}
}
