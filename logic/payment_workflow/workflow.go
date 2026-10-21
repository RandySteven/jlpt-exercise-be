package payment_workflow

import (
	"context"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var nonRetryableErrorTypes = []string{}

const (
	checkPlanActivityName             = "check_plan_activity"
	persistPaymentActivityName        = "persist_payment_activity"
	createSubscriptionActivityName    = "create_subscription_activity"
	midtransCallActivityName          = "midtrans_call_activity"
	updatePaymentStatusActivityName   = "update_payment_status_activity"
	sendSubscriptionEmailActivityName = "send_subscription_email_activity"

	sgNoNeed = ``

	paymentWorkflowName = "payment_workflow"
)

type (
	paymentWorkflow struct {
		workflow       temporal_client.WorkflowExecution
		temporalClient temporal_client.Temporal

		paymentRepository      repositories.PaymentRepository
		subscriptionRepository repositories.SubscriptionRepository
		planRepository         repositories.SubscriptionPlanRepository

		midtransClient midtrans_client.MidtransClient
	}

	PaymentWorkflow interface {
		CreateSubscription(ctx context.Context, request *requests.CreateSubscriptionRequest) (*responses.CreateSubscriptionResponse, error)
		PaymentConfirmation(ctx context.Context, request *requests.PaymentConfirmationRequest) (*responses.PaymentConfirmationResponse, error)
	}
)

func (p *paymentWorkflow) registerWorkflowAndActivities() {
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

	p.workflow.AddTransitionActivityWithOptions(checkPlanActivityName, sgNoNeed, p.checkPlan, activityOption, persistPaymentActivityName)
	p.workflow.AddTransitionActivityWithOptions(persistPaymentActivityName, sgNoNeed, p.persistPayment, activityOption, createSubscriptionActivityName)
	p.workflow.AddTransitionActivityWithOptions(createSubscriptionActivityName, sgNoNeed, p.createSubscription, activityOption, midtransCallActivityName)
	p.workflow.AddTransitionActivityWithOptions(midtransCallActivityName, sgNoNeed, p.midtransCall, activityOption, updatePaymentStatusActivityName)

	p.workflow.RegisterWorkflow(paymentWorkflowName, p.payment)
}

func (p *paymentWorkflow) CreateSubscription(ctx context.Context, request *requests.CreateSubscriptionRequest) (*responses.CreateSubscriptionResponse, error) {
	return nil, nil
}

func (p *paymentWorkflow) PaymentConfirmation(ctx context.Context, request *requests.PaymentConfirmationRequest) (*responses.PaymentConfirmationResponse, error) {
	return nil, nil
}
