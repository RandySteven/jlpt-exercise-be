package payment_workflow

import (
	"context"
	"fmt"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
	smtp_client "github.com/RandySteven/jlpt-exercise-be/pkg/smtp"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/topics"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var nonRetryableErrorTypes = []string{paymentInputErrorType, paymentRejectedErrorType}

var (
	wfInit    = temporal_client.WorkflowInit
	wfExecute = temporal_client.ExecuteWorkflow
	qryExec   = temporal_client.QueryWorkflowData
)

const (
	checkPlanActivityName             = "check_plan_activity"
	persistPaymentActivityName        = "persist_payment_activity"
	createSubscriptionActivityName    = "create_subscription_activity"
	midtransCallActivityName          = "midtrans_call_activity"
	updatePaymentStatusActivityName   = "update_payment_status_activity"
	sendSubscriptionEmailActivityName = "send_subscription_email_activity"
	publishPaymentActivityName        = "publish_payment_activity"
	planExistActivityName             = "plan_exist_activity"
	paymentFailedActivityName         = "payment_failed_activity"

	sgNoNeed             = ``
	sgCreateSubscription = `create_subscription`
	sgConfirmPayment     = `confirm_payment`
	sgCreatePayment      = `create_payment`
	sgMidtransCall       = `midtrans_call`

	paymentWorkflowName = "payment_workflow"

	queryPaymentWorkflow = "query_payment_workflow"

	paymentClientWaitTimeout  = 30 * time.Second
	paymentClientPollInterval = 150 * time.Millisecond
)

type (
	paymentWorkflow struct {
		workflow       temporal_client.WorkflowExecution
		temporalClient temporal_client.Temporal

		paymentRepository      repositories.PaymentRepository
		subscriptionRepository repositories.SubscriptionRepository
		planRepository         repositories.SubscriptionPlanRepository
		userRepository         repositories.UserRepository
		onboardingRepository   repositories.OnboardingRepository

		midtransClient midtrans_client.MidtransClient
		paymentTopic   topics.PaymentTopic
		smtpClient     smtp_client.SmtpClient
	}

	PaymentWorkflow interface {
		CreateSubscription(ctx context.Context, request *requests.CreateSubscriptionRequest) (*responses.CreateSubscriptionResponse, *apperror.CustomError)
		ContinueSubscription(ctx context.Context, request *requests.CreateSubscriptionRequest) (*responses.CreateSubscriptionResponse, *apperror.CustomError)
		PaymentConfirmation(ctx context.Context, request *requests.PaymentConfirmationRequest) (*responses.PaymentConfirmationResponse, *apperror.CustomError)
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

	p.workflow.AddTransitionActivityWithOptions(checkPlanActivityName, sgCreateSubscription, p.checkPlan, activityOption, createSubscriptionActivityName, planExistActivityName, persistPaymentActivityName)
	p.workflow.AddTransitionActivityWithOptions(planExistActivityName, sgNoNeed, p.planExist, activityOption)
	p.workflow.AddTransitionActivityWithOptions(createSubscriptionActivityName, sgNoNeed, p.createSubscription, activityOption, persistPaymentActivityName)
	p.workflow.AddTransitionActivityWithOptions(persistPaymentActivityName, sgCreatePayment, p.persistPayment, activityOption, midtransCallActivityName)
	p.workflow.AddTransitionActivityWithOptions(midtransCallActivityName, sgMidtransCall, p.midtransCall, activityOption, updatePaymentStatusActivityName, paymentFailedActivityName)

	//payment failed
	p.workflow.AddTransitionActivityWithOptions(paymentFailedActivityName, sgNoNeed, p.paymentFailed, activityOption)

	p.workflow.AddTransitionActivityWithOptions(updatePaymentStatusActivityName, sgConfirmPayment, p.confirmPayment, activityOption, publishPaymentActivityName)
	p.workflow.AddTransitionActivityWithOptions(publishPaymentActivityName, sgNoNeed, p.publishPayment, activityOption, sendSubscriptionEmailActivityName)
	p.workflow.AddTransitionActivityWithOptions(sendSubscriptionEmailActivityName, sgNoNeed, p.sendSubscriptionEmail, activityOption, sgNoNeed)

	p.workflow.RegisterWorkflow(paymentWorkflowName, p.payment)
}

func (p *paymentWorkflow) CreateSubscription(ctx context.Context, request *requests.CreateSubscriptionRequest) (*responses.CreateSubscriptionResponse, *apperror.CustomError) {
	if request == nil || request.IdempotencyKey == "" || request.PlanID == 0 {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, `subscription request is incomplete`, nil)
	}
	userID, ok := ctx.Value(enums.UserID).(uint64)
	if !ok || userID == 0 {
		return nil, apperror.NewCustomError(apperror.ErrUnauthorized, "user is missing", fmt.Errorf("user id is missing"))
	}

	workflowID := paymentWorkflowID(request.IdempotencyKey)
	startWorkflowOption := &temporal_client.StartWorkflowOptions{WorkflowID: workflowID}
	weRun, err := p.workflow.StartWorkflow(ctx, *startWorkflowOption, p.payment, request, userID)
	if err != nil {
		if alreadyStarted(err) {
			if signalErr := p.workflow.SignalWorkflow(ctx, workflowID, "", sgCreateSubscription, true); signalErr != nil {
				return nil, apperror.NewCustomError(apperror.ErrInternalServer, `failed to resume subscription workflow`, signalErr)
			}
			return p.waitForSubscriptionResponse(ctx, workflowID, "")
		}
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with workflow`, err)
	}

	if err = p.workflow.SignalWorkflow(ctx, weRun.GetID(), weRun.GetRunID(), sgCreateSubscription, true); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `failed to start subscription workflow`, err)
	}
	return p.waitForSubscriptionResponse(ctx, weRun.GetID(), weRun.GetRunID())
}

func (p *paymentWorkflow) PaymentConfirmation(ctx context.Context, request *requests.PaymentConfirmationRequest) (*responses.PaymentConfirmationResponse, *apperror.CustomError) {
	if request == nil || request.IdempotencyKey == "" || request.PlanID == 0 {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, `payment confirmation request is incomplete`, fmt.Errorf("payment confirmation request is incomplete"))
	}

	workflowID := paymentWorkflowID(request.IdempotencyKey)
	waitCtx, cancel := context.WithTimeout(ctx, paymentClientWaitTimeout)
	defer cancel()
	executionData, err := qryExec(waitCtx, p.temporalClient, queryPaymentWorkflow, workflowID, "", &ExecutionData{})
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `payment workflow is not ready`, err)
	}
	current, ok := executionData.(*ExecutionData)
	if !ok || current.Request == nil || current.Request.PlanID != request.PlanID {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, `payment confirmation does not match the subscription`, fmt.Errorf("plan id does not match"))
	}
	redirectURL := ""
	if current.MidtransTransaction != nil {
		redirectURL = current.MidtransTransaction.RedirectURL
	}
	if redirectURL == "" {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, `payment is not awaiting confirmation`, fmt.Errorf("redirect url is missing"))
	}
	if current.Subscription != nil && current.Subscription.Status == subscriptionStatusActive {
		return &responses.PaymentConfirmationResponse{RedirectURL: redirectURL}, nil
	}

	if err = p.workflow.SignalWorkflow(ctx, workflowID, current.RunID, sgConfirmPayment, true); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `payment confirmation failed`, err)
	}
	var result *responses.CreateSubscriptionResponse
	if err = p.workflow.GetWorkflowResult(ctx, workflowID, current.RunID, &result); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `payment confirmation did not complete`, err)
	}
	return &responses.PaymentConfirmationResponse{RedirectURL: redirectURL}, nil
}

func (p *paymentWorkflow) ContinueSubscription(ctx context.Context, request *requests.CreateSubscriptionRequest) (*responses.CreateSubscriptionResponse, *apperror.CustomError) {
	if request == nil || request.IdempotencyKey == "" || request.PlanID == 0 {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, `continue subscription request is incomplete`, fmt.Errorf("continue subscription request is incomplete"))
	}

	workflowID := paymentWorkflowID(request.IdempotencyKey)
	waitCtx, cancel := context.WithTimeout(ctx, paymentClientWaitTimeout)
	defer cancel()
	executionData, err := qryExec(waitCtx, p.temporalClient, queryPaymentWorkflow, workflowID, "", &ExecutionData{})
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `payment workflow is not ready`, err)
	}
	current, ok := executionData.(*ExecutionData)
	if !ok || current.Request == nil || current.Request.PlanID != request.PlanID {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, `continue subscription does not match the subscription`, fmt.Errorf("plan id does not match"))
	}
	if current.Subscription != nil && current.Subscription.Status == subscriptionStatusActive {
		return current.Response, nil
	}

	if err = p.workflow.SignalWorkflow(ctx, workflowID, current.RunID, sgCreatePayment, true); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `continue subscription failed`, err)
	}
	return p.waitForPaymentRedirect(ctx, workflowID, current.RunID)
}

func (p *paymentWorkflow) waitForPaymentRedirect(ctx context.Context, workflowID, runID string) (*responses.CreateSubscriptionResponse, *apperror.CustomError) {
	waitCtx, cancel := context.WithTimeout(ctx, paymentClientWaitTimeout)
	defer cancel()
	go p.cancelIfPaymentFailed(waitCtx, cancel, workflowID, runID)

	ticker := time.NewTicker(paymentClientPollInterval)
	defer ticker.Stop()

	for {
		executionData, err := qryExec(waitCtx, p.temporalClient, queryPaymentWorkflow, workflowID, runID, &ExecutionData{})
		if err == nil {
			current, ok := executionData.(*ExecutionData)
			if ok && current != nil && current.Response != nil && current.Response.RedirectURL != "" {
				return current.Response, nil
			}
		}

		select {
		case <-waitCtx.Done():
			if failedErr := p.paymentWorkflowFailure(workflowID, runID); failedErr != nil {
				return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with get result`, failedErr)
			}
			return nil, apperror.NewCustomError(apperror.ErrInternalServer, `payment was not initiated`, waitCtx.Err())
		case <-ticker.C:
		}
	}
}

func (p *paymentWorkflow) waitForSubscriptionResponse(ctx context.Context, workflowID, runID string) (*responses.CreateSubscriptionResponse, *apperror.CustomError) {
	waitCtx, cancel := context.WithTimeout(ctx, paymentClientWaitTimeout)
	defer cancel()
	go p.cancelIfPaymentFailed(waitCtx, cancel, workflowID, runID)

	executionData, err := qryExec(waitCtx, p.temporalClient, queryPaymentWorkflow, workflowID, runID, &ExecutionData{})
	if err != nil {
		if failedErr := p.paymentWorkflowFailure(workflowID, runID); failedErr != nil {
			return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with get result`, failedErr)
		}
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with get result`, err)
	}
	current, ok := executionData.(*ExecutionData)
	if !ok || !queryWorkflowResponseReady(current) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with convert result`, fmt.Errorf("subscription response is missing"))
	}
	return current.Response, nil
}

func (p *paymentWorkflow) cancelIfPaymentFailed(ctx context.Context, cancel context.CancelFunc, workflowID, runID string) {
	ticker := time.NewTicker(paymentClientPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			status, err := p.workflow.GetWorkflowStatus(ctx, workflowID, runID)
			if err != nil {
				continue
			}
			if status == temporal_client.StatusFailed || status == temporal_client.StatusRejected {
				cancel()
				return
			}
		}
	}
}

func (p *paymentWorkflow) paymentWorkflowFailure(workflowID, runID string) error {
	statusCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	status, err := p.workflow.GetWorkflowStatus(statusCtx, workflowID, runID)
	if err != nil {
		return nil
	}
	if status == temporal_client.StatusFailed || status == temporal_client.StatusRejected {
		return fmt.Errorf("payment workflow %s", status)
	}
	return nil
}

func NewPaymentWorkflow(
	workflow temporal_client.WorkflowExecution,
	temporal temporal_client.Temporal,
	paymentRepository repositories.PaymentRepository,
	subscriptionRepository repositories.SubscriptionRepository,
	planRepository repositories.SubscriptionPlanRepository,
	userRepository repositories.UserRepository,
	onboardingRepository repositories.OnboardingRepository,
	midtransClient midtrans_client.MidtransClient,
	paymentTopic topics.PaymentTopic,
	smtpClient smtp_client.SmtpClient,
) PaymentWorkflow {
	pw := &paymentWorkflow{
		workflow:               workflow,
		temporalClient:         temporal,
		paymentRepository:      paymentRepository,
		subscriptionRepository: subscriptionRepository,
		planRepository:         planRepository,
		userRepository:         userRepository,
		onboardingRepository:   onboardingRepository,
		midtransClient:         midtransClient,
		paymentTopic:           paymentTopic,
		smtpClient:             smtpClient,
	}
	pw.registerWorkflowAndActivities()
	return pw
}
