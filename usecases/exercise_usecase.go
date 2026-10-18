package usecases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/caches"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/logic/exercise_workflow"
	deepseek_client "github.com/RandySteven/jlpt-exercise-be/pkg/deepseek"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/topics"
	"github.com/jackc/pgx/v5/pgconn"
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
	if request == nil {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, "generate exercise request is nil", fmt.Errorf("generate exercise request is nil"))
	}

	cached, err := e.exerciseCache.GetExerciseByIdempotencyKey(ctx, request.IdempotencyKey)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to read generated exercise", err)
	}
	if cached != nil {
		return cached, nil
	}

	userID, ok := ctx.Value(enums.UserID).(uint64)
	if !ok || userID == 0 {
		return nil, apperror.NewCustomError(apperror.ErrUnauthorized, "user is missing", fmt.Errorf("user id is missing"))
	}

	latest, err := e.exerciseRepository.FindLatestByUserID(ctx, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find latest exercise", err)
	}
	if latest != nil && enums.ExerciseInProgress(latest.Status) {
		if latest.ReferenceID == request.IdempotencyKey {
			return exerciseResponseFromModel(latest), nil
		}
		return nil, apperror.NewCustomError(apperror.ErrConflict, "exercise is still in progress", fmt.Errorf("user %d exercise %d status %s", userID, latest.ID, latest.Status))
	}

	claimed, err := e.exerciseRepository.Save(ctx, &models.Exercise{
		NLevel:      request.JLPTNLevel,
		UserID:      userID,
		ReferenceID: request.IdempotencyKey,
		Status:      enums.InGeneratedStatus,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, apperror.NewCustomError(apperror.ErrConflict, "exercise is still in progress", err)
		}
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to claim exercise", err)
	}

	response, customErr = e.exerciseWorkflow.GenerateExercise(ctx, request)
	if customErr != nil {
		if errors.Is(customErr, exercise_workflow.ErrWorkflowNotStarted) {
			_, _ = e.exerciseRepository.UpdateStatusIfCurrent(ctx, claimed.ID, enums.InGeneratedStatus, enums.FailedStatus, customErr.Error())
		}
		return nil, customErr
	}
	if response != nil && response.ID == 0 {
		response.ID = claimed.ID
	}
	return response, nil
}

func exerciseResponseFromModel(exercise *models.Exercise) *responses.GenerateExerciseResponse {
	return &responses.GenerateExerciseResponse{
		ID:        exercise.ID,
		Status:    exercise.Status,
		JLPTLevel: strconv.Itoa(exercise.NLevel),
		Exercise:  exercise.Exercise,
		CreatedAt: exercise.CreatedAt,
		UpdatedAt: exercise.UpdatedAt,
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (e *exerciseUsecase) GetExercise(ctx context.Context, id uint64) (response *responses.GenerateExerciseResponse, customErr *apperror.CustomError) {
	return nil, nil
}

func (e *exerciseUsecase) SubmitExercise(ctx context.Context, request *requests.SubmitExerciseRequest) (response *responses.SubmitExerciseResponse, customErr *apperror.CustomError) {
	return nil, nil
}
