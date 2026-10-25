package refund_workflow

import "context"

func (r *refundWorkflow) updatePaymentStatus(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	return execData, nil
}
