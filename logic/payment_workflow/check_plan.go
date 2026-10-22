package payment_workflow

import (
	"context"
	"database/sql"
	"errors"

	temporal_client "github.com/RandySteven/go-cook/temporal"
)

func (p *paymentWorkflow) checkPlan(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Request == nil {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "subscription request is missing")
	}
	if execData.UserID == 0 || execData.Request.PlanID == 0 || execData.Request.IdempotencyKey == "" {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "subscription request is incomplete")
	}

	user, err := p.userRepository.FindByID(ctx, execData.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "user does not exist")
		}
		return nil, err
	}
	onboarding, err := p.onboardingRepository.FindByID(ctx, user.OnboardingID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "onboarding does not exist")
		}
		return nil, err
	}
	if onboarding.Email == "" {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "customer email is missing")
	}
	execData.CustomerEmail = onboarding.Email
	execData.CustomerName = onboarding.Name

	plan, err := p.planRepository.FindByID(ctx, execData.Request.PlanID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "plan does not exist")
		}
		return nil, err
	}
	if plan.Price <= 0 || plan.Currency == "" {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "plan price is invalid")
	}
	execData.Plan = plan

	current, err := p.subscriptionRepository.FindCurrentByUserID(ctx, execData.UserID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if current != nil && current.ID != 0 {
		execData.Subscription = current
		execData.Status = current.Status
		execData.setSubscriptionResponse(current.ID)
		execData.SetActivity(planExistActivityName)
		return execData, nil
	}

	execData.Status = subscriptionStatusPending
	execData.SetActivity(persistPaymentActivityName)
	return execData, nil
}

func (p *paymentWorkflow) planExist(_ context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Subscription == nil {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "subscription is missing")
	}
	execData.setSubscriptionResponse(execData.Subscription.ID)
	execData.Status = execData.Subscription.Status
	return execData, nil
}
