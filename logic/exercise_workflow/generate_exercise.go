package exercise_workflow

import (
	"context"
	"fmt"
	"strconv"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"go.temporal.io/sdk/workflow"
)

const (
	exerciseStatusProcessing = "processing"
	exerciseStatusCompleted  = "completed"
)

const queryGenerateExerciseResponse = "GenerateExerciseResponse"

func (e *exerciseWorkflow) generateExercise(ctx workflow.Context, request *requests.GenerateExerciseRequest, userID uint64) (response *responses.GenerateExerciseResponse, err error) {
	if request == nil {
		return nil, fmt.Errorf("generate exercise request is nil")
	}
	if userID == 0 {
		return nil, fmt.Errorf("user id is missing")
	}

	info := workflow.GetInfo(ctx)
	now := workflow.Now(ctx)
	executionData := &ExecutionData{
		UserRequest:    request,
		UserID:         userID,
		IdempotencyKey: request.IdempotencyKey,
		WorkflowID:     info.WorkflowExecution.ID,
		RunID:          info.WorkflowExecution.RunID,
		GenerateExercise: &responses.GenerateExerciseResponse{
			Status:    exerciseStatusProcessing,
			JLPTLevel: strconv.Itoa(request.JLPTNLevel),
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	executionData.SetActivity(readPromptTxtActivity)

	executionDataResult, err := wfInit(ctx, e.workflow, queryGenerateExerciseResponse, executionData, func(data temporal_client.ExecutionData) error {
		current, ok := data.(*ExecutionData)
		if !ok || !exerciseResponseReady(current) {
			return fmt.Errorf("generated exercise is not ready")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	executionData, ok := executionDataResult.(*ExecutionData)
	if !ok {
		return nil, fmt.Errorf("there is wrong with convert result")
	}

	if _, err = wfExecute(ctx, e.workflow, executionData); err != nil {
		return nil, err
	}
	if executionData.GenerateExercise == nil {
		return nil, fmt.Errorf("generated exercise is missing")
	}
	return executionData.GenerateExercise, nil
}

func exerciseResponseReady(current *ExecutionData) bool {
	if current == nil || current.GenerateExercise == nil {
		return false
	}
	return current.GenerateExercise.Status == exerciseStatusProcessing || len(current.GenerateExercise.Exercise) > 0
}

func (e *exerciseWorkflow) cacheProcessingExercise(ctx context.Context, execData *ExecutionData) error {
	if execData == nil || execData.GenerateExercise == nil || execData.GenerateExercise.Status != exerciseStatusProcessing {
		return nil
	}
	key := execData.IdempotencyKey
	if key == "" && execData.UserRequest != nil {
		key = execData.UserRequest.IdempotencyKey
	}
	if key == "" {
		return nil
	}
	return e.exerciseCache.SetExerciseByIdempotencyKey(ctx, key, execData.GenerateExercise)
}

func (e *exerciseWorkflow) signalExerciseEvent(ctx context.Context, execData *ExecutionData, signalName string) error {
	if execData == nil || execData.WorkflowID == "" {
		return fmt.Errorf("workflow id is missing")
	}
	return e.workflow.SignalWorkflow(ctx, execData.WorkflowID, execData.RunID, signalName, true)
}
