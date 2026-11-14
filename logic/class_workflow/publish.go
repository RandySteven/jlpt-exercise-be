package class_workflow

import (
	"context"
	"errors"
)

func (c *classWorkflow) publishClass(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil {
		return nil, errors.New("execution data is missing")
	}
	execData.SetActivity(pushNotificationActivityName)
	return execData, nil
}
