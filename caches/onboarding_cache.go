package caches

import (
	"context"
	"fmt"
	"time"

	redis_client "github.com/RandySteven/go-cook/redis"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

const (
	onboardingLoginAttemptKey     = `onboarding:login:email:%s:attempt`
	onboardingLoginAttemptTTL     = 1 * time.Hour
	onboardingLoginKey            = `onboarding:login:email:%s:token`
	onboardingLoginTTL            = 1 * time.Hour
	onboardingRegisterResponse    = `onboarding:register:email:%s:response`
	onboardingRegisterResponseTTL = 15 * time.Minute
)

type (
	OnboardingCache interface {
		Get(ctx context.Context, key string) (val interface{}, err error)
		Set(ctx context.Context, key string, val interface{}) (err error)
		GetLoginAttempt(ctx context.Context, key string) (attempt int, err error)
		SetLoginAttempt(ctx context.Context, key string, attempt int) error
		SetRegisterResponse(ctx context.Context, email string, response *responses.RegisterResponse) error
		GetRegisterResponse(ctx context.Context, email string) (response *responses.RegisterResponse, err error)
	}

	onboardingCache struct {
		cache redis_client.Redis
	}
)

func (o *onboardingCache) Get(ctx context.Context, key string) (val interface{}, err error) {
	value, err := o.cache.Get(ctx, fmt.Sprintf(onboardingLoginKey, key))
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (o *onboardingCache) GetLoginAttempt(ctx context.Context, key string) (attempt int, err error) {
	value, err := o.cache.Get(ctx, fmt.Sprintf(onboardingLoginAttemptKey, key))
	if err != nil {
		return 0, err
	}
	switch n := value.(type) {
	case float64:
		return int(n), nil
	case int:
		return n, nil
	default:
		return 0, fmt.Errorf("unexpected login attempt value %T", value)
	}
}

func (o *onboardingCache) Set(ctx context.Context, key string, val interface{}) (err error) {
	_, err = o.cache.Set(ctx, fmt.Sprintf(onboardingLoginKey, key), val, onboardingLoginTTL)
	return err
}

func (o *onboardingCache) SetLoginAttempt(ctx context.Context, key string, attempt int) error {
	_, err := o.cache.Set(ctx, fmt.Sprintf(onboardingLoginAttemptKey, key), attempt, onboardingLoginAttemptTTL)
	return err
}

func (o *onboardingCache) SetRegisterResponse(ctx context.Context, email string, response *responses.RegisterResponse) error {
	_, err := o.cache.Set(ctx, fmt.Sprintf(onboardingRegisterResponse, email), response, onboardingRegisterResponseTTL)
	return err
}

func (o *onboardingCache) GetRegisterResponse(ctx context.Context, email string) (response *responses.RegisterResponse, err error) {
	value, err := o.cache.Get(ctx, fmt.Sprintf(onboardingRegisterResponse, email))
	if err != nil {
		return nil, err
	}
	return value.(*responses.RegisterResponse), nil
}

func newOnboardingCache(cache redis_client.Redis) OnboardingCache {
	return &onboardingCache{
		cache: cache,
	}
}
