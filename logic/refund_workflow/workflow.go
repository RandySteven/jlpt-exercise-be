package refund_workflow

import (
	"context"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var (
	checkPaymentExistingActivity = "check_payment_existing"
	createRefundActivity         = "create_refund"
	refundPaymentActivity        = "refund_payment"
	midtransRefundActivity       = "midtrans_refund"
	refundPaymentSuccessActivity = "refund_payment_success"
	refundPaymentFailedActivity  = "refund_payment_failed"

	workflowName = `refund_workflow`

	nonRetryableErrorTypes = []string{
		`temporal:non_retryable_error`,
		`temporal:non_retryable_error_type`,
		`temporal:non_retryable_error_type_not_found`,
		`temporal:non_retryable_error_type_not_found`,
	}

	sgNoNeed         = ``
	sgMidtransRefund = `midtrans_refund`
)

type (
	refundWorkflow struct {
		workflow       temporal_client.WorkflowExecution
		temporalClient temporal_client.Temporal
	}

	RefundWorkflow interface {
		Refund(ctx context.Context)
	}
)

func (r *refundWorkflow) registerActivitiesAndWorkflow() {
	activityOption := &workflow.ActivityOptions{
		StartToCloseTimeout: 25 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:        5,
			InitialInterval:        1 * time.Minute,
			BackoffCoefficient:     2,
			MaximumInterval:        5 * time.Minute,
			NonRetryableErrorTypes: nonRetryableErrorTypes,
		},
	}

	r.workflow.AddTransitionActivityWithOptions(checkPaymentExistingActivity, sgNoNeed, r.checkPayment, activityOption, createRefundActivity)
	r.workflow.AddTransitionActivityWithOptions(createRefundActivity, sgNoNeed, r.initiateRefund, activityOption, midtransRefundActivity)
	r.workflow.AddTransitionActivityWithOptions(midtransRefundActivity, sgMidtransRefund, r.midtransCall, activityOption, refundPaymentSuccessActivity, refundPaymentFailedActivity)

	r.workflow.RegisterWorkflow(workflowName, r.refund)
}
