package payment_workflow

import (
	"context"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
)

func (p *paymentWorkflow) confirmPayment(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Payment == nil || execData.Subscription == nil {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "payment confirmation is missing")
	}
	plan := execData.Plan
	if plan == nil {
		loaded, err := p.planRepository.FindByID(ctx, execData.Subscription.PlanID)
		if err != nil {
			return nil, err
		}
		plan = loaded
		execData.Plan = loaded
	}

	execData.Payment.Status = paymentStatusPaid
	payment, err := p.paymentRepository.Update(ctx, execData.Payment)
	if err != nil {
		return nil, err
	}
	execData.Payment = payment

	start := time.Now().UTC()
	end, err := subscriptionEnd(start, plan)
	if err != nil {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, err.Error())
	}
	execData.Subscription.Status = subscriptionStatusActive
	execData.Subscription.StartDate = start
	execData.Subscription.EndDate = end
	subscription, err := p.subscriptionRepository.Update(ctx, execData.Subscription)
	if err != nil {
		return nil, err
	}

	execData.Subscription = subscription
	execData.Status = subscriptionStatusActive
	execData.setSubscriptionResponse(subscription.ID)
	execData.SetActivity(publishPaymentActivityName)
	return execData, nil
}
