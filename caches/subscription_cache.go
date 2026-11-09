package caches

import (
	"context"
	"encoding/json"
	"fmt"

	redis_client "github.com/RandySteven/go-cook/redis"
	"github.com/RandySteven/jlpt-exercise-be/configs"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

const (
	userSubscriptionDetailKey = "user_subscription_detail:%s"
)

type (
	SubscriptionCache interface {
		SetUserSubscriptionDetail(ctx context.Context, key string, value *responses.GetUserSubscriptionDetailResponse) error
		GetUserSubscriptionDetail(ctx context.Context, key string) (*responses.GetUserSubscriptionDetailResponse, error)
	}

	subscriptionCache struct {
		redis       redis_client.Redis
		cacheConfig map[string]configs.CacheConfig
	}
)

func (s *subscriptionCache) SetUserSubscriptionDetail(ctx context.Context, key string, value *responses.GetUserSubscriptionDetailResponse) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = s.redis.Set(ctx, fmt.Sprintf(userSubscriptionDetailKey, key), string(payload), getTTL(s.cacheConfig, "user_subscription_detail"))
	return err
}

func (s *subscriptionCache) GetUserSubscriptionDetail(ctx context.Context, key string) (*responses.GetUserSubscriptionDetailResponse, error) {
	value, err := s.redis.Get(ctx, fmt.Sprintf(userSubscriptionDetailKey, key))
	if err != nil {
		return nil, err
	}
	return decodeCacheValue[responses.GetUserSubscriptionDetailResponse](value)
}

func newSubscriptionCache(redis redis_client.Redis, cacheConfig map[string]configs.CacheConfig) SubscriptionCache {
	return &subscriptionCache{redis: redis, cacheConfig: cacheConfig}
}
