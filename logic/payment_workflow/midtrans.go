package payment_workflow

import (
	"context"

	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
)

func (p *paymentWorkflow) midtransCall(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	response, err := p.midtransClient.CreateTransaction(ctx, &midtrans_client.TransactionRequest{})
	if err != nil {
		return nil, err
	}
	execData.MidtransTransaction = response
	return execData, nil
}
