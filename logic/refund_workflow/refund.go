package refund_workflow

import (
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"go.temporal.io/sdk/workflow"
)

func (r *refundWorkflow) refund(ctx workflow.Context, request *requests.RefundRequest) (*responses.RefundResponse, error) {
	return nil, nil
}
