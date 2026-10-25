package refund_workflow

import "context"

func (r *refundWorkflow) checkPayment(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	return execData, nil
}
