package payment_workflow

import "context"

func (p *paymentWorkflow) midtransCall(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	return execData, nil
}
