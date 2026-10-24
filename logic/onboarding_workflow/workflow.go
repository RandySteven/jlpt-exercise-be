package onboarding_workflow

import (
	"context"
	"fmt"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/caches"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	smtp_client "github.com/RandySteven/jlpt-exercise-be/pkg/smtp"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var (
	wfInit     = temporal_client.WorkflowInit
	wfQueryGet = temporal_client.QueryWorkflowData
	wfExecute  = temporal_client.ExecuteWorkflow
)

const (
	onboardingWorkflowExecution = `onboarding_workflow_execution`

	persistOnboardingRequestActivity = `persist_onboarding_request_activity`
	registerUserActivity             = `register_user_activity`
	updateOnboardingStatusActivity   = `update_onboarding_status_activity`
	sendActivationEmailActivity      = `send_activation_email_activity`

	sgNoEvent            = ``
	sgActivatedUserEvent = `activated_user_signal`

	queryRegisterResponse = `RegisterResponse`

	registerClientWaitTimeout  = 30 * time.Second
	registerClientPollInterval = 150 * time.Millisecond
)

var nonRetryableErrorTypes = []string{onboardingInputErrorType}

type (
	onboardingWorkflow struct {
		workflow temporal_client.WorkflowExecution
		temporal temporal_client.Temporal

		onboardingRepository repositories.OnboardingRepository
		userRepository       repositories.UserRepository
		onboardingCache      caches.OnboardingCache
		smtpClient           smtp_client.SmtpClient
	}

	OnboardingWorkflow interface {
		OnboardingRegisterWorkflow(ctx context.Context, request *requests.RegisterRequest) (response *responses.RegisterResponse, customErr *apperror.CustomError)
		ActivateUserRegisterWorkflow(ctx context.Context, request *requests.ActivateRequest) (response *responses.ActivateResponse, customErr *apperror.CustomError)
	}
)

func (o *onboardingWorkflow) registerWorkflowAndActivities() {
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

	o.workflow.AddTransitionActivityWithOptions(persistOnboardingRequestActivity, sgNoEvent, o.persistRequest, activityOption, registerUserActivity)
	o.workflow.AddTransitionActivityWithOptions(registerUserActivity, sgNoEvent, o.registerUser, activityOption, sendActivationEmailActivity)
	o.workflow.AddTransitionActivityWithOptions(sendActivationEmailActivity, sgNoEvent, o.sendActivationEmail, activityOption, updateOnboardingStatusActivity)
	o.workflow.AddTransitionActivityWithOptions(updateOnboardingStatusActivity, sgActivatedUserEvent, o.updateOnboardingStatus, activityOption)

	o.workflow.RegisterWorkflow(onboardingWorkflowExecution, o.register)
}

func (o *onboardingWorkflow) OnboardingRegisterWorkflow(ctx context.Context, request *requests.RegisterRequest) (response *responses.RegisterResponse, customErr *apperror.CustomError) {
	workflowID := fmt.Sprintf("%s-%s", onboardingWorkflowExecution, ctx.Value(enums.RequestID))
	startWorkflowOption := &temporal_client.StartWorkflowOptions{
		WorkflowID: workflowID,
	}
	weRun, err := o.workflow.StartWorkflow(ctx, *startWorkflowOption, o.register, workflowID, request)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with workflow `, err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, registerClientWaitTimeout)
	defer cancel()
	go o.cancelIfRegisterFailed(waitCtx, cancel, weRun.GetID(), weRun.GetRunID())

	executionData, err := wfQueryGet(waitCtx, o.temporal, queryRegisterResponse, weRun.GetID(), weRun.GetRunID(), &ExecutionData{})
	if err != nil {
		if failedErr := o.registerWorkflowFailure(weRun.GetID(), weRun.GetRunID()); failedErr != nil {
			return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with get result`, failedErr)
		}
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with get result`, err)
	}

	executionDataResult, ok := executionData.(*ExecutionData)
	if !ok || executionDataResult.Response == nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with convert result`, fmt.Errorf("register response is missing"))
	}
	return executionDataResult.Response, nil
}

func (o *onboardingWorkflow) cancelIfRegisterFailed(ctx context.Context, cancel context.CancelFunc, workflowID, runID string) {
	ticker := time.NewTicker(registerClientPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			status, err := o.workflow.GetWorkflowStatus(ctx, workflowID, runID)
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

func (o *onboardingWorkflow) registerWorkflowFailure(workflowID, runID string) error {
	statusCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	status, err := o.workflow.GetWorkflowStatus(statusCtx, workflowID, runID)
	if err != nil {
		return nil
	}
	if status == temporal_client.StatusFailed || status == temporal_client.StatusRejected {
		return fmt.Errorf("onboarding workflow %s", status)
	}
	return nil
}

func (o *onboardingWorkflow) ActivateUserRegisterWorkflow(ctx context.Context, request *requests.ActivateRequest) (response *responses.ActivateResponse, customErr *apperror.CustomError) {
	obj, err := decodeActivationToken(request.Token)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `invalid token`, err)
	}

	if err := o.workflow.SignalWorkflow(ctx, obj.WorkflowID, obj.RunID, sgActivatedUserEvent, true); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `activation failed`, err)
	}

	var registerResult *responses.RegisterResponse
	if err := o.workflow.GetWorkflowResult(ctx, obj.WorkflowID, obj.RunID, &registerResult); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `activation did not complete`, err)
	}

	return &responses.ActivateResponse{Status: "success", Activate: true}, nil
}

func NewOnboardingWorkflow(
	workflow temporal_client.WorkflowExecution,
	temporal temporal_client.Temporal,
	onboardingRepository repositories.OnboardingRepository,
	userRepository repositories.UserRepository,
	onboardingCache caches.OnboardingCache,
	smtpClient smtp_client.SmtpClient,
) *onboardingWorkflow {
	ow := &onboardingWorkflow{
		workflow:             workflow,
		temporal:             temporal,
		onboardingRepository: onboardingRepository,
		userRepository:       userRepository,
		onboardingCache:      onboardingCache,
		smtpClient:           smtpClient,
	}
	ow.registerWorkflowAndActivities()
	return ow
}
