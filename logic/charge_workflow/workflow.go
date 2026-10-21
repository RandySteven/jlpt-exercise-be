package charge_workflow

import (
	"context"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	chargeActivityName  = "charge_activity"
	captureActivityName = "capture_activity"
	publishActivityName = "publish_activity"
	chargeWorkflowName  = "ChargeWorkflow"
)

var nonRetryableErrorTypes = []string{}

type (
	chargeWorkflow struct {
		workflow       temporal_client.WorkflowExecution
		temporalClient temporal_client.Temporal

	}

	ChargeWorkflow interface {
		Charge(ctx context.Context, request *requests.ChargeSubscriptionRequest) (*responses.ChargeSubscriptionResponse, *apperror.CustomError)
		Capture(ctx context.Context, request *requests.CaptureSubscriptionRequest) (*responses.CaptureSubscriptionResponse, *apperror.CustomError)
	}
)

func (c *chargeWorkflow) chargeWorkflowAndActivities() {
	_ = &workflow.ActivityOptions{
		StartToCloseTimeout: 25 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:        5,
			InitialInterval:        1 * time.Minute,
			BackoffCoefficient:     2,
			MaximumInterval:        5 * time.Minute,
			NonRetryableErrorTypes: nonRetryableErrorTypes,
		},
	}

	c.workflow.RegisterWorkflow(chargeWorkflowName, c.charge)
}

func (c *chargeWorkflow) Charge(ctx context.Context, request *requests.ChargeSubscriptionRequest) (*responses.ChargeSubscriptionResponse, *apperror.CustomError) {
	return nil, nil
}

func (c *chargeWorkflow) Capture(ctx context.Context, request *requests.CaptureSubscriptionRequest) (*responses.CaptureSubscriptionResponse, *apperror.CustomError) {
	return nil, nil
}
