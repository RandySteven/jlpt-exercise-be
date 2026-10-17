package exercise_workflow

import "context"

func (e *exerciseWorkflow) persistUserRequest(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	return execData, nil
}

func (e *exerciseWorkflow) saveExercise(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	return execData, nil
}
