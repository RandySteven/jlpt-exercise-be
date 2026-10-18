package usecases

import (
	"context"
	"errors"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/caches"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/logic/exercise_workflow"
	deepseek_client "github.com/RandySteven/jlpt-exercise-be/pkg/deepseek"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/topics"
	"github.com/redis/go-redis/v9"
)

type (
	ExerciseUsecase interface {
		GenerateExercise(ctx context.Context, request *requests.GenerateExerciseRequest) (response *responses.GenerateExerciseResponse, customErr *apperror.CustomError)
		GetExercise(ctx context.Context, id uint64) (response *responses.GenerateExerciseResponse, customErr *apperror.CustomError)
		SubmitExercise(ctx context.Context, request *requests.SubmitExerciseRequest) (response *responses.SubmitExerciseResponse, customErr *apperror.CustomError)
	}
)

type (
	exerciseUsecase struct {
		deepSeek           deepseek_client.DeepSeek
		exerciseWorkflow   exercise_workflow.ExerciseWorkflow
		exerciseRepository repositories.ExerciseRepository
		userRepository     repositories.UserRepository
		exerciseTopic      topics.ExerciseTopic
		exerciseCache      caches.ExerciseCache
	}
)

func newExerciseUsecase(
	deepSeek deepseek_client.DeepSeek,
	temporal temporal_client.Temporal,
	exerciseRepository repositories.ExerciseRepository,
	userRepository repositories.UserRepository,
	exerciseCache caches.ExerciseCache,
	exerciseTopic topics.ExerciseTopic,
) ExerciseUsecase {
	exerciseWorkflow := exercise_workflow.NewExerciseWorkflow(
		temporal_client.NewWorkflowExecution(temporal),
		temporal,
		exerciseRepository,
		userRepository,
		exerciseCache,
		deepSeek,
	)
	return &exerciseUsecase{
		deepSeek:           deepSeek,
		exerciseWorkflow:   exerciseWorkflow,
		exerciseRepository: exerciseRepository,
		userRepository:     userRepository,
		exerciseCache:      exerciseCache,
		exerciseTopic:      exerciseTopic,
	}
}

func (e *exerciseUsecase) GenerateExercise(ctx context.Context, request *requests.GenerateExerciseRequest) (response *responses.GenerateExerciseResponse, customErr *apperror.CustomError) {
	getGeneratedExercise, err := e.exerciseCache.GetExerciseByIdempotencyKey(ctx, request.IdempotencyKey)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, err.Error(), err)
	}

	if getGeneratedExercise != nil {
		return getGeneratedExercise, nil
	}

	response, err = e.exerciseWorkflow.GenerateExercise(ctx, request)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, err.Error(), err)
	}

	return response, nil
}

func (e *exerciseUsecase) GetExercise(ctx context.Context, id uint64) (response *responses.GenerateExerciseResponse, customErr *apperror.CustomError) {
	return nil, nil
}

func (e *exerciseUsecase) SubmitExercise(ctx context.Context, request *requests.SubmitExerciseRequest) (response *responses.SubmitExerciseResponse, customErr *apperror.CustomError) {
	return nil, nil
}
