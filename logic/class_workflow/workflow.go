package class_workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
	smtp_client "github.com/RandySteven/jlpt-exercise-be/pkg/smtp"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var (
	wfInit      = temporal_client.WorkflowInit
	wfExecute   = temporal_client.ExecuteWorkflow
	wfGetResult = temporal_client.GetWorkflowData
	wfQuery     = temporal_client.QueryWorkflowData
)

var (
	persistClassActivityName               = "persist_class_activity"
	validateUserActivityName               = "validate_user_activity"
	assignUserToClassActivityName          = "assign_user_to_class_activity"
	initiatedPaymentForStudentActivityName = "initiated_payment_for_student_activity"
	initiateMidtransForStudentActivityName = "initiate_midtrans_for_student_activity"
	publishClassActivityName               = "publish_class_activity"
	pushNotificationActivityName           = "push_notification_activity"
	finishWorkflowActivityName             = "finish_workflow_activity"
	rejectStudentActivityName              = "reject_student_activity"

	sgNoNeed         = ``
	sgCreateClass    = `sg_create_class`
	sgAddUserToClass = `sg_add_user_to_class`

	workflowName = "assign_sensei_class_workflow"
)

const queryAssignSenseiClassResponse = "AssignSenseiClassResponse"

const queryAssignSenseiClassResponseTimeout = 10 * time.Second

type (
	classWorkflow struct {
		workflow       temporal_client.WorkflowExecution
		temporalClient temporal_client.Temporal

		classRepository          repositories.ClassRepository
		classCatalogueRepository repositories.ClassCatalogueRepository
		classCategoryRepository  repositories.ClassCategoryRepository
		classUserRepository      repositories.ClassUserRepository
		senseiRepository         repositories.SenseiRepository
		userRepository           repositories.UserRepository

		paymentRepository    repositories.PaymentRepository
		onboardingRepository repositories.OnboardingRepository
		midtransClient       midtrans_client.MidtransClient
		smtpClient           smtp_client.SmtpClient

		transaction repositories.Transaction
	}

	ClassWorkflow interface {
		AssignSenseiClassWorkflow(ctx context.Context, request *requests.AssignSenseiClassRequest, index uint64) (response *responses.AssignSenseiClassResponse, customErr *apperror.CustomError)
		AppendClassUser(ctx context.Context, request *requests.AssignSenseiClassRequest, index uint64) (response *responses.AssignSenseiClassResponse, customErr *apperror.CustomError)
	}
)

func (c *classWorkflow) registerActivitiesAndWorkflow() {
	ao := workflow.ActivityOptions{
		TaskQueue:           "assign_sensei_class_task_queue",
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}

	c.workflow.AddTransitionActivityWithOptions(persistClassActivityName, sgCreateClass, c.persistClass, &ao, validateUserActivityName)
	c.workflow.AddTransitionActivityWithOptions(validateUserActivityName, sgAddUserToClass, c.validateUser, &ao, assignUserToClassActivityName, rejectStudentActivityName, validateUserActivityName, finishWorkflowActivityName)
	c.workflow.AddTransitionActivityWithOptions(rejectStudentActivityName, sgNoNeed, c.rejectStudent, &ao, validateUserActivityName, finishWorkflowActivityName)
	c.workflow.AddTransitionActivityWithOptions(assignUserToClassActivityName, sgNoNeed, c.assignUserToClass, &ao, initiatedPaymentForStudentActivityName)

	c.workflow.AddTransitionActivityWithOptions(initiatedPaymentForStudentActivityName, sgNoNeed, c.initiatedPaymentForStudent, &ao, initiateMidtransForStudentActivityName)
	c.workflow.AddTransitionActivityWithOptions(initiateMidtransForStudentActivityName, sgNoNeed, c.initiateMidtransForStudent, &ao, publishClassActivityName)
	c.workflow.AddTransitionActivityWithOptions(publishClassActivityName, sgNoNeed, c.publishClass, &ao, pushNotificationActivityName)
	c.workflow.AddTransitionActivityWithOptions(pushNotificationActivityName, sgNoNeed, c.pushNotification, &ao, validateUserActivityName, finishWorkflowActivityName)
	c.workflow.AddTransitionActivityWithOptions(finishWorkflowActivityName, sgNoNeed, c.finishWorkflow, &ao)

	c.workflow.RegisterWorkflow(workflowName, c.createClass)
}

func (c *classWorkflow) AssignSenseiClassWorkflow(ctx context.Context, request *requests.AssignSenseiClassRequest, index uint64) (response *responses.AssignSenseiClassResponse, customErr *apperror.CustomError) {
	if request == nil || request.ReferenceID == "" || len(request.StudentIDs) == 0 || index >= uint64(len(request.StudentIDs)) {
		return nil, apperror.InvalidRequestErr()
	}

	workflowID := classWorkflowID(request.ReferenceID)
	options := temporal_client.StartWorkflowOptions{
		WorkflowID: workflowID,
	}

	wfRun, err := c.workflow.StartWorkflow(ctx, options, c.createClass, request)
	if err != nil {
		if !alreadyStarted(err) {
			return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with workflow`, err)
		}
		if signalErr := c.workflow.SignalWorkflow(ctx, workflowID, "", sgCreateClass, true); signalErr != nil {
			return nil, apperror.NewCustomError(apperror.ErrInternalServer, `failed to resume class workflow`, signalErr)
		}
		return c.AppendClassUser(ctx, request, index)
	}

	if err = c.workflow.SignalWorkflow(ctx, wfRun.GetID(), wfRun.GetRunID(), sgCreateClass, true); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `failed to start class workflow`, err)
	}
	if err = c.workflow.SignalWorkflow(ctx, wfRun.GetID(), wfRun.GetRunID(), sgAddUserToClass, addStudentSignal{RequestStudentIndex: index}); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `failed to add student to class workflow`, err)
	}
	return c.waitForAssignedStudent(ctx, wfRun.GetID(), wfRun.GetRunID())
}

func (c *classWorkflow) AppendClassUser(ctx context.Context, request *requests.AssignSenseiClassRequest, index uint64) (response *responses.AssignSenseiClassResponse, customErr *apperror.CustomError) {
	if request == nil || request.ReferenceID == "" || len(request.StudentIDs) == 0 || index >= uint64(len(request.StudentIDs)) {
		return nil, apperror.InvalidRequestErr()
	}

	workflowID := classWorkflowID(request.ReferenceID)
	if err := c.workflow.SignalWorkflow(ctx, workflowID, "", sgAddUserToClass, addStudentSignal{RequestStudentIndex: index}); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `failed to add student to class workflow`, err)
	}
	return c.waitForAssignedStudent(ctx, workflowID, "")
}

func (c *classWorkflow) waitForAssignedStudent(ctx context.Context, workflowID, runID string) (*responses.AssignSenseiClassResponse, *apperror.CustomError) {
	waitCtx, cancel := context.WithTimeout(ctx, queryAssignSenseiClassResponseTimeout)
	defer cancel()

	executionData, err := wfQuery(waitCtx, c.temporalClient, queryAssignSenseiClassResponse, workflowID, runID, &ExecutionData{})
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with get result`, err)
	}
	current, ok := executionData.(*ExecutionData)
	if !ok || !queryWorkflowResponseReady(current) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with convert result`, fmt.Errorf("assign sensei class response is missing"))
	}
	return current.Response, nil
}

func NewClassWorkflow(
	workflow temporal_client.WorkflowExecution,
	temporalClient temporal_client.Temporal,
	classRepository repositories.ClassRepository,
	classCatalogueRepository repositories.ClassCatalogueRepository,
	classUserRepository repositories.ClassUserRepository,
	senseiRepository repositories.SenseiRepository,
	userRepository repositories.UserRepository,
	onboardingRepository repositories.OnboardingRepository,
	paymentRepository repositories.PaymentRepository,
	transaction repositories.Transaction,
	midtransClient midtrans_client.MidtransClient,
	smtpClient smtp_client.SmtpClient,
) ClassWorkflow {
	cw := &classWorkflow{
		workflow:                 workflow,
		temporalClient:           temporalClient,
		classRepository:          classRepository,
		classCatalogueRepository: classCatalogueRepository,
		classUserRepository:      classUserRepository,
		senseiRepository:         senseiRepository,
		userRepository:           userRepository,
		onboardingRepository:     onboardingRepository,
		paymentRepository:        paymentRepository,
		transaction:              transaction,
		midtransClient:           midtransClient,
		smtpClient:               smtpClient,
	}

	cw.registerActivitiesAndWorkflow()

	return cw
}

func classWorkflowID(referenceID string) string {
	return fmt.Sprintf("%s-%s", workflowName, referenceID)
}

func alreadyStarted(err error) bool {
	var target *serviceerror.WorkflowExecutionAlreadyStarted
	return errors.As(err, &target)
}
