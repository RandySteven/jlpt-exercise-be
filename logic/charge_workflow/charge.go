package charge_workflow

import (
	"context"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

func (c *chargeWorkflow) charge(ctx context.Context, request *requests.ChargeSubscriptionRequest) (*responses.ChargeSubscriptionResponse, *apperror.CustomError) {
	return nil, nil
}
