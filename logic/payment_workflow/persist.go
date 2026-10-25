package payment_workflow

import (
	"context"
	"database/sql"
	"errors"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
)

func (p *paymentWorkflow) persistPayment(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Request == nil || execData.Plan == nil || execData.Subscription == nil || execData.Subscription.ID == 0 {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "payment input is missing")
	}

	existing, err := p.paymentRepository.FindByReferenceID(ctx, execData.Request.IdempotencyKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existing != nil && existing.ID != 0 {
		if existing.UserID != execData.UserID {
			return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "payment reference belongs to another user")
		}
		execData.Payment = existing
		if existing.Status == paymentStatusPaid {
			return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "payment already paid")
		}
		execData.SetActivity(midtransCallActivityName)
		if _, err = p.signalMidtransCall(ctx, execData); err != nil {
			return nil, err
		}
		return execData, nil
	}

	metadata, err := encodePaymentMetadata(paymentMetadata{
		PlanID:   execData.Plan.ID,
		PlanName: execData.Plan.Name,
	})
	if err != nil {
		return nil, err
	}
	payment, err := p.paymentRepository.Save(ctx, &models.Payment{
		ReferenceID:    execData.Request.IdempotencyKey,
		SubscriptionID: execData.Subscription.ID,
		UserID:         execData.UserID,
		Amount:         execData.Plan.Price,
		Currency:       execData.Plan.Currency,
		Status:         paymentStatusPending,
		Metadata:       metadata,
	})
	if err != nil {
		if isUniqueViolation(err) {
			existing, findErr := p.paymentRepository.FindByReferenceID(ctx, execData.Request.IdempotencyKey)
			if findErr != nil {
				return nil, findErr
			}
			if existing.UserID != execData.UserID {
				return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "payment reference belongs to another user")
			}
			execData.Payment = existing
			execData.SetActivity(midtransCallActivityName)
			if _, signalErr := p.signalMidtransCall(ctx, execData); signalErr != nil {
				return nil, signalErr
			}
			return execData, nil
		}
		return nil, err
	}

	execData.Payment = payment
	execData.Status = paymentStatusPending
	execData.SetActivity(midtransCallActivityName)
	if _, err = p.signalMidtransCall(ctx, execData); err != nil {
		return nil, err
	}
	return execData, nil
}

func (p *paymentWorkflow) createSubscription(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Request == nil || execData.Plan == nil || execData.UserID == 0 {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "subscription input is missing")
	}

	existing, err := p.subscriptionRepository.FindCurrentByUserID(ctx, execData.UserID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existing != nil && existing.ID != 0 {
		execData.Subscription = existing
		execData.fillSubscriptionResponse(existing)
		if existing.Status == subscriptionStatusActive {
			execData.SetActivity(planExistActivityName)
			return execData, nil
		}
		execData.SetActivity(persistPaymentActivityName)
		return execData, nil
	}

	start := time.Now().UTC()
	end, err := subscriptionEnd(start, execData.Plan)
	if err != nil {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, err.Error())
	}

	subscription, err := p.subscriptionRepository.Save(ctx, &models.Subscription{
		UserID:      execData.UserID,
		ReferenceID: execData.Request.IdempotencyKey,
		PlanID:      execData.Plan.ID,
		StartDate:   start,
		EndDate:     end,
		Status:      subscriptionStatusPending,
	})
	if err != nil {
		return nil, err
	}

	execData.Subscription = subscription
	execData.Status = subscriptionStatusPending
	execData.fillSubscriptionResponse(subscription)
	execData.SetActivity(persistPaymentActivityName)
	if _, err = p.signalCreatePayment(ctx, execData); err != nil {
		return nil, err
	}
	return execData, nil
}

func (p *paymentWorkflow) signalMidtransCall(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Subscription == nil || execData.Subscription.ID == 0 {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "subscription input is missing")
	}

	if err := p.workflow.SignalWorkflow(ctx, execData.WorkflowID, execData.RunID, sgMidtransCall, true); err != nil {
		return nil, err
	}
	return execData, nil
}

func (p *paymentWorkflow) signalCreatePayment(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Subscription == nil || execData.Subscription.ID == 0 {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "subscription input is missing")
	}

	if err := p.workflow.SignalWorkflow(ctx, execData.WorkflowID, execData.RunID, sgCreatePayment, true); err != nil {
		return nil, err
	}
	return execData, nil
}
