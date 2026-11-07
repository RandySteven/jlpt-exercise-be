package class_workflow

import (
	"context"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var (
	persistClassActivityName      = "persist_class_activity"
	validateUserActivityName      = "validate_user_activity"
	assignUserToClassActivityName = "assign_user_to_class_activity"
	publishClassActivityName      = "publish_class_activity"

	sgNoNeed         = ``
	sgCreateClass    = `sg_create_class`
	sgAddUserToClass = `sg_add_user_to_class`
	sgPublishClass   = `sg_publish_class`
	sgAppendStudent  = `sg_append_student`

	workflowName = "assign_sensei_class_workflow"
)

type (
	classWorkflow struct {
		workflow       temporal_client.WorkflowExecution
		temporalClient temporal_client.Temporal

		classRepository     repositories.ClassRepository
		classUserRepository repositories.ClassUserRepository
		senseiRepository    repositories.SenseiRepository
		userRepository      repositories.UserRepository

		transaction repositories.Transaction
	}

	ClassWorkflow interface {
		AssignSenseiClassWorkflow(ctx context.Context, request *requests.AssignSenseiClassRequest) (response *responses.AssignSenseiClassResponse, customErr *apperror.CustomError)
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
	c.workflow.AddTransitionActivityWithOptions(validateUserActivityName, sgAddUserToClass, c.validateUser, &ao, assignUserToClassActivityName)
	c.workflow.AddTransitionActivityWithOptions(assignUserToClassActivityName, sgPublishClass, c.assignUserToClass, &ao, publishClassActivityName)

	c.workflow.RegisterWorkflow(workflowName, c.createClass)
}

func (c *classWorkflow) AssignSenseiClassWorkflow(ctx context.Context, request *requests.AssignSenseiClassRequest) (response *responses.AssignSenseiClassResponse, customErr *apperror.CustomError) {
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
