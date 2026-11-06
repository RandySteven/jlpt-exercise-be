package class_workflow

import (
	"context"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
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
)

type (
	classWorkflow struct {
		workflow       temporal_client.WorkflowExecution
		temporalClient temporal_client.Temporal

		classRepository     repositories.ClassRepository
		classUserRepository repositories.ClassUserRepository
		senseiRepository    repositories.SenseiRepository
		userRepository      repositories.UserRepository
	}

	ClassWorkflow interface {
		AssignSenseiClassWorkflow(ctx context.Context, request *requests.AssignSenseiClassRequest) (response *responses.AssignSenseiClassResponse, customErr *apperror.CustomError)
	}
)

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
	return &classWorkflow{
		workflow:            workflow,
		temporalClient:      temporalClient,
		classRepository:     classRepository,
		classUserRepository: classUserRepository,
		senseiRepository:    senseiRepository,
		userRepository:      userRepository,
	}
}
