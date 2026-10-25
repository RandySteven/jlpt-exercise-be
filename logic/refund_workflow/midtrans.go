package refund_workflow

import "context"

func (r *refundWorkflow) midtransCall(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	return execData, nil
}
