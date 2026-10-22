package payment_workflow

import (
	"context"
	"strconv"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
)

func (p *paymentWorkflow) midtransCall(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Payment == nil || execData.Plan == nil || execData.Subscription == nil {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "midtrans input is missing")
	}
	if execData.Payment.ReferenceID == "" || execData.Payment.Amount <= 0 {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "payment amount is invalid")
	}

	if execData.Payment.ID != 0 {
		current, err := p.paymentRepository.FindByID(ctx, execData.Payment.ID)
		if err != nil {
			return nil, err
		}
		if current != nil {
			execData.Payment = current
		}
	}

	meta := decodePaymentMetadata(execData.Payment.Metadata)
	if meta.Token != "" && meta.RedirectURL != "" {
		execData.MidtransTransaction = &midtrans_client.TransactionResponse{
			Token:       meta.Token,
			RedirectURL: meta.RedirectURL,
		}
		execData.setSubscriptionResponse(execData.Subscription.ID)
		execData.SetActivity(updatePaymentStatusActivityName)
		return execData, nil
	}

	itemName := execData.Plan.Name
	if itemName == "" {
		itemName = "Subscription"
	}
	request := &midtrans_client.TransactionRequest{
		TransactionDetails: midtrans_client.TransactionDetails{
			OrderID:     execData.Payment.ReferenceID,
			GrossAmount: int64(execData.Payment.Amount),
		},
		ItemDetails: []midtrans_client.ItemDetails{{
			ID:       strconv.FormatUint(execData.Plan.ID, 10),
			Price:    int64(execData.Plan.Price),
			Quantity: 1,
			Name:     itemName,
		}},
		CustomField1: strconv.FormatUint(execData.Subscription.ID, 10),
		CustomField2: strconv.FormatUint(execData.Payment.ID, 10),
	}
	if execData.CustomerEmail != "" || execData.CustomerName != "" {
		request.CustomerDetails = &midtrans_client.CustomerDetails{
			FirstName: execData.CustomerName,
			Email:     execData.CustomerEmail,
		}
	}

	response, err := p.midtransClient.CreateTransaction(ctx, request)
	if err != nil {
		if rejectedMidtrans(err) {
			execData.SetActivity(paymentFailedActivityName)
			return execData, nil
		}
		return nil, err
	}
	if response == nil || response.Token == "" || response.RedirectURL == "" {
		return nil, temporal_client.NewNonRetryableError(paymentRejectedErrorType, "midtrans transaction response is incomplete")
	}

	meta.Token = response.Token
	meta.RedirectURL = response.RedirectURL
	raw, err := encodePaymentMetadata(meta)
	if err != nil {
		return nil, err
	}
	execData.Payment.Metadata = raw
	updated, err := p.paymentRepository.Update(ctx, execData.Payment)
	if err != nil {
		return nil, err
	}

	execData.Payment = updated
	execData.MidtransTransaction = response
	execData.setSubscriptionResponse(execData.Subscription.ID)
	execData.SetActivity(updatePaymentStatusActivityName)
	return execData, nil
}

func (p *paymentWorkflow) paymentFailed(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData != nil && execData.Payment != nil && execData.Payment.ID != 0 {
		execData.Payment.Status = paymentStatusFailed
		updated, err := p.paymentRepository.Update(ctx, execData.Payment)
		if err != nil {
			return nil, err
		}
		execData.Payment = updated
	}
	if execData != nil && execData.Subscription != nil && execData.Subscription.ID != 0 {
		execData.Subscription.Status = subscriptionStatusFailed
		updated, err := p.subscriptionRepository.Update(ctx, execData.Subscription)
		if err != nil {
			return nil, err
		}
		execData.Subscription = updated
	}
	if execData != nil {
		execData.Status = paymentStatusFailed
	}
	return nil, temporal_client.NewNonRetryableError(paymentRejectedErrorType, "midtrans transaction was rejected")
}
