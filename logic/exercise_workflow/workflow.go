package exercise_workflow

import (
	"context"
	"fmt"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var (
	//activity names
	persistUserRequestActivity = "persist_user_request_activity"
	readPromptTxtActivity      = "read_prompt_txt_activity"
	generateExerciseActivity   = "generate_exercise_activity"
	saveExerciseActivity       = "save_exercise_activity"
	publishActivity            = "publish_activity"
	completeActivity           = "complete_activity"

	failedActivity = "failed_activity"

	//sg event
	sgNoNeed     = ``
	sgPersist    = `sg_persist_user_request`
	sgProcessing = `sg_processing_generation`
	sgCompleted  = `sg_completed_generation`

	workflowName = "exercise_workflow"

	generateExerciseClientWaitTimeout = 10 * time.Second
)

var (
	wfInit      = temporal_client.WorkflowInit
	wfExecute   = temporal_client.ExecuteWorkflow
	wfGetResult = temporal_client.GetWorkflowData
	wfQuery     = temporal_client.QueryWorkflowData
)

var (
	nonRetryableErrorTypes = []string{}
)

type (
	exerciseWorkflow struct {
		workflow       temporal_client.WorkflowExecution
		temporalClient temporal_client.Temporal
	}

	ExerciseWorkflow interface {
		GenerateExercise(ctx context.Context, request *requests.GenerateExerciseRequest) (response *responses.GenerateExerciseResponse, customErr *apperror.CustomError)
	}
)

func (e *exerciseWorkflow) registerActivitiesAndWorkflow() {
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

	e.workflow.AddTransitionActivityWithOptions(persistUserRequestActivity, sgNoNeed, e.persistUserRequest, activityOption, readPromptTxtActivity)
	e.workflow.AddTransitionActivityWithOptions(readPromptTxtActivity, sgNoNeed, e.readPromptTxt, nil, generateExerciseActivity, failedActivity)
	e.workflow.AddTransitionActivityWithOptions(generateExerciseActivity, sgProcessing, e.generate, nil, saveExerciseActivity, failedActivity)

	e.workflow.AddTransitionActivityWithOptions(saveExerciseActivity, sgCompleted, e.saveExercise, nil, publishActivity, failedActivity)
	e.workflow.AddTransitionActivityWithOptions(publishActivity, sgNoNeed, e.publish, nil, completeActivity)

	e.workflow.RegisterWorkflow(workflowName, e.generateExercise)
}

func (e *exerciseWorkflow) GenerateExercise(ctx context.Context, request *requests.GenerateExerciseRequest) (response *responses.GenerateExerciseResponse, customErr *apperror.CustomError) {
	workflowID := fmt.Sprintf("%s-%s", workflowName, request.IdempotencyKey)

	startWorkflowOptions := temporal_client.StartWorkflowOptions{
		WorkflowID: workflowID,
	}

	_, err := e.workflow.StartWorkflow(ctx, startWorkflowOptions, e.generateExercise, request)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with workflow `, err)
	}

	return nil, nil
}
