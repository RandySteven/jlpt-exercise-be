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
	PaymentCache interface {
		SetUserSubscriptionDetail(ctx context.Context, key string, value *responses.GetUserSubscriptionDetailResponse) error
		GetUserSubscriptionDetail(ctx context.Context, key string) (*responses.GetUserSubscriptionDetailResponse, error)
	}

	paymentCache struct {
		redis       redis_client.Redis
		cacheConfig map[string]configs.CacheConfig
	}
)

func (p *paymentCache) SetUserSubscriptionDetail(ctx context.Context, key string, value *responses.GetUserSubscriptionDetailResponse) error {
	json, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = p.redis.Set(ctx, fmt.Sprintf(userSubscriptionDetailKey, key), string(json), getTTL(p.cacheConfig, "user_subscription_detail"))
	return err
}

func (p *paymentCache) GetUserSubscriptionDetail(ctx context.Context, key string) (*responses.GetUserSubscriptionDetailResponse, error) {
	value, err := p.redis.Get(ctx, fmt.Sprintf(userSubscriptionDetailKey, key))
	if err != nil {
		return nil, err
	}
	return decodeCacheValue[responses.GetUserSubscriptionDetailResponse](value)
}

func newPaymentCache(redis redis_client.Redis, cacheConfig map[string]configs.CacheConfig) PaymentCache {
	return &paymentCache{redis: redis, cacheConfig: cacheConfig}
}
