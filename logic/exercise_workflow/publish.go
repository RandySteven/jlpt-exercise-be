package exercise_workflow

import "context"

func (e *exerciseWorkflow) publish(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	execData.SetActivity(completeActivity)
	return execData, nil
}
