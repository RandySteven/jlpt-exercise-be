package class_workflow

import (
	"context"
	"fmt"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
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
	confirmClassPaymentActivityName        = "confirm_class_payment_activity"
	finishWorkflowActivityName             = "finish_workflow_activity"
	rejectStudentActivityName              = "reject_student_activity"

	sgNoNeed              = ``
	sgCreateClass         = `sg_create_class`
	sgAddUserToClass      = `sg_add_user_to_class`
	sgConfirmClassPayment = `sg_confirm_class_payment`

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

		paymentRepository repositories.PaymentRepository
		midtransClient    midtrans_client.MidtransClient

		transaction repositories.Transaction
	}

	ClassWorkflow interface {
		AssignSenseiClassWorkflow(ctx context.Context, request *requests.AssignSenseiClassRequest, index uint64) (response *responses.AssignSenseiClassResponse, customErr *apperror.CustomError)
		ConfirmClassPayment(ctx context.Context, request *requests.ConfirmClassPaymentRequest) (response *responses.ConfirmClassPaymentResponse, customErr *apperror.CustomError)
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
	c.workflow.AddTransitionActivityWithOptions(validateUserActivityName, sgAddUserToClass, c.validateUser, &ao, assignUserToClassActivityName, rejectStudentActivityName)
	c.workflow.AddTransitionActivityWithOptions(assignUserToClassActivityName, sgNoNeed, c.assignUserToClass, &ao, publishClassActivityName)

	c.workflow.AddTransitionActivityWithOptions(initiatedPaymentForStudentActivityName, sgNoNeed, c.initiatedPaymentForStudent, &ao, initiateMidtransForStudentActivityName)
	c.workflow.AddTransitionActivityWithOptions(initiateMidtransForStudentActivityName, sgNoNeed, c.initiateMidtransForStudent, &ao, publishClassActivityName)
	c.workflow.AddTransitionActivityWithOptions(publishClassActivityName, sgNoNeed, c.publishClass, &ao, pushNotificationActivityName)

	c.workflow.AddTransitionActivityWithOptions(pushNotificationActivityName, sgNoNeed, c.pushNotification, &ao, confirmClassPaymentActivityName, finishWorkflowActivityName)

	c.workflow.AddTransitionActivityWithOptions(confirmClassPaymentActivityName, sgConfirmClassPayment, c.confirmClassPayment, &ao, pushNotificationActivityName)

	c.workflow.RegisterWorkflow(workflowName, c.createClass)
}

func (c *classWorkflow) AssignSenseiClassWorkflow(ctx context.Context, request *requests.AssignSenseiClassRequest, index uint64) (response *responses.AssignSenseiClassResponse, customErr *apperror.CustomError) {
	if request == nil || len(request.StudentIDs) == 0 {
		return nil, apperror.InvalidRequestErr()
	}

	if index >= uint64(len(request.StudentIDs)) {
		return nil, apperror.InvalidRequestErr()
	}

	workflowID := fmt.Sprintf("%s-%s", workflowName, request.ReferenceID)

	options := temporal_client.StartWorkflowOptions{
		WorkflowID: workflowID,
	}

	wfRun, err := c.workflow.StartWorkflow(ctx, options, c.createClass, request, index)
	if err != nil {
		return nil, apperror.InternalServerErr()
	}

	waitCtx, cancel := context.WithTimeout(ctx, queryAssignSenseiClassResponseTimeout)
	defer cancel()

	executionData, err := wfQuery(waitCtx, c.temporalClient, queryAssignSenseiClassResponse, wfRun.GetID(), wfRun.GetRunID(), &ExecutionData{})
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with get result`, err)
	}
	executionDataResult, ok := executionData.(*ExecutionData)
	if !ok || !exerciseResponseReady(executionDataResult) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with convert result`, fmt.Errorf("generated exercise is missing"))
	}

	return executionDataResult.Response, nil
}

func (c *classWorkflow) ConfirmClassPayment(ctx context.Context, request *requests.ConfirmClassPaymentRequest) (response *responses.ConfirmClassPaymentResponse, customErr *apperror.CustomError) {
	panic("unimplemented")
}

func newClassWorkflow(
	workflow temporal_client.WorkflowExecution,
	temporalClient temporal_client.Temporal,
	classRepository repositories.ClassRepository,
	classUserRepository repositories.ClassUserRepository,
	senseiRepository repositories.SenseiRepository,
	userRepository repositories.UserRepository,
) ClassWorkflow {
	cw := &classWorkflow{
		workflow:            workflow,
		temporalClient:      temporalClient,
		classRepository:     classRepository,
		classUserRepository: classUserRepository,
		senseiRepository:    senseiRepository,
		userRepository:      userRepository,
	}

	cw.registerActivitiesAndWorkflow()

	return cw
}

func exerciseResponseReady(current *ExecutionData) bool {
	if current == nil || current.Response == nil {
		return false
	}
	return current.Response.Status == `PROCESSING` || current.Response.Status == `COMPLETED`
}
