package caches

import (
	"context"
	"fmt"
	"time"

	redis_client "github.com/RandySteven/go-cook/redis"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

const (
	exerciseKey            = `exercise:id:%d`
	exerciseTTL            = 1 * time.Hour
	exerciseIdempotencyKey = `exercise:idempotency:%s`
	exerciseIdempotencyTTL = 1 * time.Hour
	exerciseSubmitKey      = `exercise:submit:id:%d`
	exerciseSubmitTTL      = 1 * time.Hour
)

type (
	ExerciseCache interface {
		SetExercise(ctx context.Context, id uint64, response *responses.GenerateExerciseResponse) error
		GetExercise(ctx context.Context, id uint64) (response *responses.GenerateExerciseResponse, err error)
		SetExerciseByIdempotencyKey(ctx context.Context, key string, response *responses.GenerateExerciseResponse) error
		GetExerciseByIdempotencyKey(ctx context.Context, key string) (response *responses.GenerateExerciseResponse, err error)
		DeleteExercise(ctx context.Context, id uint64) error
		DeleteExerciseByIdempotencyKey(ctx context.Context, key string) error
		SetSubmitExercise(ctx context.Context, id uint64, response *responses.SubmitExerciseResponse) error
		GetSubmitExercise(ctx context.Context, id uint64) (response *responses.SubmitExerciseResponse, err error)
		DeleteSubmitExercise(ctx context.Context, id uint64) error
	}

	exerciseCache struct {
		cache redis_client.Redis
	}
)

func (e *exerciseCache) SetExercise(ctx context.Context, id uint64, response *responses.GenerateExerciseResponse) error {
	_, err := e.cache.Set(ctx, fmt.Sprintf(exerciseKey, id), response, exerciseTTL)
	return err
}

func (e *exerciseCache) GetExercise(ctx context.Context, id uint64) (response *responses.GenerateExerciseResponse, err error) {
	value, err := e.cache.Get(ctx, fmt.Sprintf(exerciseKey, id))
	if err != nil {
		return nil, err
	}
	return decodeCacheValue[responses.GenerateExerciseResponse](value)
}

func (e *exerciseCache) SetExerciseByIdempotencyKey(ctx context.Context, key string, response *responses.GenerateExerciseResponse) error {
	_, err := e.cache.Set(ctx, fmt.Sprintf(exerciseIdempotencyKey, key), response, exerciseIdempotencyTTL)
	return err
}

func (e *exerciseCache) DeleteExercise(ctx context.Context, id uint64) error {
	return e.cache.Del(ctx, fmt.Sprintf(exerciseKey, id))
}

func (e *exerciseCache) DeleteExerciseByIdempotencyKey(ctx context.Context, key string) error {
	return e.cache.Del(ctx, fmt.Sprintf(exerciseIdempotencyKey, key))
}

func (e *exerciseCache) GetExerciseByIdempotencyKey(ctx context.Context, key string) (response *responses.GenerateExerciseResponse, err error) {
	value, err := e.cache.Get(ctx, fmt.Sprintf(exerciseIdempotencyKey, key))
	if err != nil {
		return nil, err
	}
	return decodeCacheValue[responses.GenerateExerciseResponse](value)
}

func (e *exerciseCache) SetSubmitExercise(ctx context.Context, id uint64, response *responses.SubmitExerciseResponse) error {
	_, err := e.cache.Set(ctx, fmt.Sprintf(exerciseSubmitKey, id), response, exerciseSubmitTTL)
	return err
}

func (e *exerciseCache) GetSubmitExercise(ctx context.Context, id uint64) (response *responses.SubmitExerciseResponse, err error) {
	value, err := e.cache.Get(ctx, fmt.Sprintf(exerciseSubmitKey, id))
	if err != nil {
		return nil, err
	}
	return decodeCacheValue[responses.SubmitExerciseResponse](value)
}

func (e *exerciseCache) DeleteSubmitExercise(ctx context.Context, id uint64) error {
	return e.cache.Del(ctx, fmt.Sprintf(exerciseSubmitKey, id))
}

func newExerciseCache(cache redis_client.Redis) ExerciseCache {
	return &exerciseCache{
		cache: cache,
	}
}

var _ ExerciseCache = &exerciseCache{}
