package caches

import (
	"context"
	"encoding/json"
	"fmt"

	redis_client "github.com/RandySteven/go-cook/redis"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/configs"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

var (
	subscriptionPlanDetailCacheKey = "subscription_plan_detail:%d"
	subscriptionPlanListCacheKey   = "subscription_plan_list"
)

type (
	SubscriptionPlanCache interface {
		GetSubscriptionPlanDetail(ctx context.Context, id uint64) (response *responses.GetSubscriptionPlanDetailResponse, err error)
		SetSubscriptionPlanDetail(ctx context.Context, id uint64, response *responses.GetSubscriptionPlanDetailResponse) (err error)
		ListSubscriptionPlans(ctx context.Context) (response *responses.ListSubscriptionPlansResponse, err error)
		SetListSubscriptionPlans(ctx context.Context, response *responses.ListSubscriptionPlansResponse) (err error)
	}

	subscriptionPlanCache struct {
		redisClient redis_client.Redis
		cacheConfig map[string]configs.CacheConfig
	}
)

func newSubscriptionPlanCache(redisClient redis_client.Redis, cacheConfig map[string]configs.CacheConfig) SubscriptionPlanCache {
	return &subscriptionPlanCache{redisClient: redisClient, cacheConfig: cacheConfig}
}

func (c *subscriptionPlanCache) GetSubscriptionPlanDetail(ctx context.Context, id uint64) (response *responses.GetSubscriptionPlanDetailResponse, err error) {
	value, err := c.redisClient.Get(ctx, fmt.Sprintf(subscriptionPlanDetailCacheKey, id))
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to get subscription plan detail", err)
	}
	return decodeCacheValue[responses.GetSubscriptionPlanDetailResponse](value)
}

func (c *subscriptionPlanCache) SetSubscriptionPlanDetail(ctx context.Context, id uint64, response *responses.GetSubscriptionPlanDetailResponse) (err error) {
	json, err := json.Marshal(response)
	if err != nil {
		return err
	}
	tTL := getTTL(c.cacheConfig, "subscription_plan_detail")
	_, err = c.redisClient.Set(ctx, fmt.Sprintf(subscriptionPlanDetailCacheKey, id), string(json), tTL)
	return err
}

func (c *subscriptionPlanCache) ListSubscriptionPlans(ctx context.Context) (response *responses.ListSubscriptionPlansResponse, err error) {
	value, err := c.redisClient.Get(ctx, subscriptionPlanListCacheKey)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to get subscription plan list", err)
	}
	return decodeCacheValue[responses.ListSubscriptionPlansResponse](value)
}

func (c *subscriptionPlanCache) SetListSubscriptionPlans(ctx context.Context, response *responses.ListSubscriptionPlansResponse) (err error) {
	json, err := json.Marshal(response)
	if err != nil {
		return err
	}
	tTL := getTTL(c.cacheConfig, "subscription_plan_list")
	_, err = c.redisClient.Set(ctx, subscriptionPlanListCacheKey, string(json), tTL)
	return err
}
