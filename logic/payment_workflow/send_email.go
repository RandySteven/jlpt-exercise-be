package payment_workflow

import (
	"context"
	"fmt"
	"log"

	temporal_client "github.com/RandySteven/go-cook/temporal"
)

func (p *paymentWorkflow) sendSubscriptionEmail(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Subscription == nil {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "subscription email is missing")
	}
	if execData.CustomerEmail == "" {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "customer email is missing")
	}
	if p.smtpClient == nil {
		return nil, temporal_client.NewNonRetryableError(paymentInputErrorType, "smtp client is missing")
	}

	planName := "your plan"
	if execData.Plan != nil && execData.Plan.Name != "" {
		planName = execData.Plan.Name
	}
	subject := "Your JLPT subscription is active"
	body := fmt.Sprintf(
		"<p>Your subscription to %s is active until %s.</p>",
		planName,
		execData.Subscription.EndDate.UTC().Format("2006-01-02"),
	)
	if err := p.smtpClient.SendEmail(execData.CustomerEmail, subject, body); err != nil {
		log.Println("failed to send subscription email", err)
	}
	return execData, nil
}
