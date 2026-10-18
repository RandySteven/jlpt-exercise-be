package exercise_workflow

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
)

const exerciseInputErrorType = "ExerciseInputError"

func (e *exerciseWorkflow) persistExerciseGeneratedResponse(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.GenerateExercise == nil {
		return nil, temporal_client.NewNonRetryableError(exerciseInputErrorType, "generated exercise is missing")
	}

	key := execData.IdempotencyKey
	if key == "" && execData.UserRequest != nil {
		key = execData.UserRequest.IdempotencyKey
	}
	if key == "" {
		return nil, temporal_client.NewNonRetryableError(exerciseInputErrorType, "idempotency key is missing")
	}

	if execData.Exercise == nil {
		return nil, temporal_client.NewNonRetryableError(exerciseInputErrorType, "exercise is missing")
	}

	updatedExercise, err := e.exerciseRepository.Update(ctx, execData.Exercise)
	if err != nil {
		log.Println("failed to update exercise", err)
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to update exercise", err)
	}

	if err = e.exerciseCache.DeleteExercise(ctx, updatedExercise.ID); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to flush exercise cache", err)
	}
	if err = e.exerciseCache.DeleteExerciseByIdempotencyKey(ctx, key); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to flush exercise cache", err)
	}

	execData.Exercise = updatedExercise
	execData.GenerateExercise = generateExerciseResponseFromExercise(updatedExercise)
	execData.IdempotencyKey = key
	if err = e.exerciseCache.SetExercise(ctx, updatedExercise.ID, execData.GenerateExercise); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to persist generated exercise", err)
	}
	if err = e.exerciseCache.SetExerciseByIdempotencyKey(ctx, key, execData.GenerateExercise); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to persist generated exercise", err)
	}

	execData.SetActivity(publishActivity)
	return execData, nil
}

func generateExerciseResponseFromExercise(exercise *models.Exercise) *responses.GenerateExerciseResponse {
	return &responses.GenerateExerciseResponse{
		ID:        exercise.ID,
		Status:    exercise.Status,
		JLPTLevel: strconv.Itoa(exercise.NLevel),
		Exercise:  exercise.Exercise,
		CreatedAt: exercise.CreatedAt,
		UpdatedAt: exercise.UpdatedAt,
	}
}

func (e *exerciseWorkflow) persistUserRequest(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.UserRequest == nil {
		return nil, temporal_client.NewNonRetryableError(exerciseInputErrorType, "user request is missing")
	}

	if execData.UserID == 0 {
		return nil, temporal_client.NewNonRetryableError(exerciseInputErrorType, "user id is missing")
	}
	user, err := e.userRepository.FindByID(ctx, execData.UserID)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find user", err)
	}

	execData.User = user

	latest, err := e.exerciseRepository.FindLatestByUserID(ctx, user.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find latest exercise", err)
	}
	if latest != nil && enums.ExerciseInProgress(latest.Status) {
		if latest.ReferenceID == execData.UserRequest.IdempotencyKey {
			execData.Exercise = latest
			return execData, nil
		}
		return nil, temporal_client.NewNonRetryableError(exerciseInputErrorType, "user already has an exercise in progress")
	}

	saved, err := e.exerciseRepository.Save(ctx, &models.Exercise{
		NLevel:      execData.UserRequest.JLPTNLevel,
		UserID:      execData.User.ID,
		ReferenceID: execData.UserRequest.IdempotencyKey,
		Status:      enums.InGeneratedStatus,
	})
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to persist user request", err)
	}

	execData.Exercise = saved
	return execData, nil
}
