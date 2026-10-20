package payment_workflow

import "context"

func (p *paymentWorkflow) persistPayment(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	return execData, nil
}
