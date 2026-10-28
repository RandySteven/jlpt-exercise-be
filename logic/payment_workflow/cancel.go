package payment_workflow

import (
	"context"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
)

func (p *paymentWorkflow) cancel(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Payment == nil || execData.Payment.ID == 0 {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "payment not found")
	}

	transaction, err := p.midtransClient.GetTransaction(ctx, execData.Subscription.ReferenceID)
	if err != nil {
		return nil, err
	}

	switch transaction.TransactionStatus {
	case enums.MidtransSettlementStatus.String(), enums.MidtransCaptureStatus.String():
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "payment already settled")
	case enums.MidtransPendingStatus.String():
		_, err := p.midtransClient.Cancel(ctx, &midtrans_client.CancelRequest{
			OrderID: execData.Subscription.ReferenceID,
		})
		if err != nil {
			return nil, err
		}
	case enums.MidtransCancelStatus.String():
		execData.SetActivity(updatePaymentStatusActivityName)
	}

	return execData, nil
}
