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
	if execData == nil || execData.Request == nil || execData.Plan == nil {
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
		execData.SetActivity(createSubscriptionActivityName)
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
		ReferenceID: execData.Request.IdempotencyKey,
		UserID:      execData.UserID,
		Amount:      execData.Plan.Price,
		Currency:    execData.Plan.Currency,
		Status:      paymentStatusPending,
		Metadata:    metadata,
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
			execData.SetActivity(createSubscriptionActivityName)
			return execData, nil
		}
		return nil, err
	}

	execData.Payment = payment
	execData.Status = paymentStatusPending
	execData.SetActivity(createSubscriptionActivityName)
	return execData, nil
}

func (p *paymentWorkflow) createSubscription(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Payment == nil || execData.Plan == nil || execData.Payment.ID == 0 {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "subscription input is missing")
	}

	existing, err := p.subscriptionRepository.FindByPaymentID(ctx, execData.Payment.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existing != nil && existing.ID != 0 {
		execData.Subscription = existing
		execData.SetActivity(midtransCallActivityName)
		return execData, nil
	}

	start := time.Now().UTC()
	end, err := subscriptionEnd(start, execData.Plan)
	if err != nil {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, err.Error())
	}

	subscription, err := p.subscriptionRepository.Save(ctx, &models.Subscription{
		UserID:    execData.UserID,
		PaymentID: execData.Payment.ID,
		PlanID:    execData.Plan.ID,
		StartDate: start,
		EndDate:   end,
		Status:    subscriptionStatusPending,
	})
	if err != nil {
		if isUniqueViolation(err) {
			existing, findErr := p.subscriptionRepository.FindByPaymentID(ctx, execData.Payment.ID)
			if findErr != nil {
				return nil, findErr
			}
			execData.Subscription = existing
			execData.SetActivity(midtransCallActivityName)
			return execData, nil
		}
		return nil, err
	}

	execData.Subscription = subscription
	execData.Status = subscriptionStatusPending
	execData.SetActivity(midtransCallActivityName)
	return execData, nil
}
