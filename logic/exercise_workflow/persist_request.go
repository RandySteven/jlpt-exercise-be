package exercise_workflow

import (
	"context"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
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

	if err := e.exerciseCache.SetExerciseByIdempotencyKey(ctx, key, execData.GenerateExercise); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to persist generated exercise", err)
	}

	execData.IdempotencyKey = key
	execData.SetActivity(publishActivity)
	return execData, nil
}
