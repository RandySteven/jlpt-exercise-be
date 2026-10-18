package exercise_workflow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
)

func (e *exerciseWorkflow) generate(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	result, err := e.deepSeekClient.ChatGenerate(ctx, execData.Prompt)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to generate text", err)
	}
	if !jsonFormatChecker(result) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to generate text", errors.New("invalid json format"))
	}
	if execData.GenerateExercise == nil {
		execData.GenerateExercise = &responses.GenerateExerciseResponse{}
	}
	execData.GenerateExercise.Status = enums.GeneratedStatus
	execData.GenerateExercise.Exercise = json.RawMessage(result)
	execData.GenerateExercise.UpdatedAt = time.Now()
	execData.SetActivity(persistUserRequestActivity)
	if err = e.signalExerciseEvent(ctx, execData, sgCompleted); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to signal completed generation", err)
	}
	return execData, nil
}

func jsonFormatChecker(str string) bool {
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(str), &jsonData)
	if err != nil {
		return false
	}
	return true
}
