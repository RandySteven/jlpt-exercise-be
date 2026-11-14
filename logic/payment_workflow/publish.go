package payment_workflow

import (
	"context"
	"log"
	"strconv"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/schemas"
)

func (p *paymentWorkflow) publishPayment(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Payment == nil || execData.Subscription == nil {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "payment publish is missing")
	}
	if p.paymentTopic == nil {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "payment topic is missing")
	}

	var userName, userEmail string
	if execData.Onboarding != nil {
		userName = execData.Onboarding.Name
		userEmail = execData.Onboarding.Email
	}

	err := p.paymentTopic.WriteMessage(ctx, &schemas.PaymentSchema{
		ID:              strconv.FormatUint(execData.Payment.ID, 10),
		Amount:          float64(execData.Payment.Amount),
		TransactionType: execData.Payment.TransactionType,
		TransactionID:   execData.Payment.TransactionID,
		UserID:          strconv.FormatUint(execData.UserID, 10),
		UserName:        userName,
		UserEmail:       userEmail,
		Metadata:        execData.Payment.Metadata,
		Currency:        execData.Payment.Currency,
		Status:          execData.Payment.Status,
		CreatedAt:       execData.Payment.CreatedAt,
		UpdatedAt:       execData.Payment.UpdatedAt,
	})
	if err != nil {
		log.Println("failed to publish payment", err)
	}

	log.Println("payment published", execData.Payment.ID)

	execData.SetActivity(sendSubscriptionEmailActivityName)
	return execData, nil
}
