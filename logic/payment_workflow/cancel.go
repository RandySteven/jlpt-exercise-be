package payment_workflow

import (
	"context"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
)

func (p *paymentWorkflow) cancel(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	transaction, err := p.midtransClient.GetTransaction(ctx, execData.Subscription.ReferenceID)
	if err != nil {
		return nil, err
	}

	switch transaction.TransactionStatus {
	case `settlement`:
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "payment already settled")
	case `pending`:
		_, err := p.midtransClient.Cancel(ctx, &midtrans_client.CancelRequest{
			OrderID: execData.Subscription.ReferenceID,
		})
		if err != nil {
			return nil, err
		}
	case `cancel`:
		execData.SetActivity(updatePaymentStatusActivityName)
	}

	return execData, nil
}
