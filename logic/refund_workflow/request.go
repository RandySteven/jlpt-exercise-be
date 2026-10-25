package refund_workflow

import "context"

func (r *refundWorkflow) initiateRefund(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	return execData, nil
}
