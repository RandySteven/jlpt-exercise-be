package payment_workflow

import (
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"go.temporal.io/sdk/workflow"
)

func (p *paymentWorkflow) payment(ctx workflow.Context, request *requests.CreateSubscriptionRequest) (*responses.CreateSubscriptionResponse, error) {
	return nil, nil
}
