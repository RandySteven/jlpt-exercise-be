package exercise_workflow

import (
	"context"
	"log"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
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

	maxAttempts := 3
	go func() {
		for i := 0; i < maxAttempts; i++ {
			updatedExercise, err := e.exerciseRepository.Update(ctx, execData.Exercise)
			if err != nil {
				log.Println("failed to update exercise", err)
				return
			}
			if updatedExercise.Status == enums.InGeneratedStatus {
				execData.Exercise = updatedExercise
				break
			}
			time.Sleep(1 * time.Second)
		}
	}()

	if err := e.exerciseCache.SetExerciseByIdempotencyKey(ctx, key, execData.GenerateExercise); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to persist generated exercise", err)
	}

	execData.IdempotencyKey = key
	execData.SetActivity(publishActivity)
	return execData, nil
}

func (e *exerciseWorkflow) persistUserRequest(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.UserRequest == nil {
		return nil, temporal_client.NewNonRetryableError(exerciseInputErrorType, "user request is missing")
	}

	userID := ctx.Value(enums.UserID).(uint64)
	user, err := e.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find user", err)
	}

	execData.User = user

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
