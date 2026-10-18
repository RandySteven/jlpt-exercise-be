package exercise_workflow

import (
	"context"
	"fmt"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/caches"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	deepseek_client "github.com/RandySteven/jlpt-exercise-be/pkg/deepseek"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	fileContentPath = "files/txt/generate_exercise.txt"
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
	nonRetryableErrorTypes = []string{exerciseInputErrorType}
)

type (
	exerciseWorkflow struct {
		workflow           temporal_client.WorkflowExecution
		temporalClient     temporal_client.Temporal
		exerciseRepository repositories.ExerciseRepository
		userRepository     repositories.UserRepository
		exerciseCache      caches.ExerciseCache
		deepSeekClient     deepseek_client.DeepSeek
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

	e.workflow.AddTransitionActivityWithOptions(readPromptTxtActivity, sgNoNeed, e.readPromptTxt, activityOption, generateExerciseActivity, failedActivity)
	e.workflow.AddTransitionActivityWithOptions(generateExerciseActivity, sgProcessing, e.generate, activityOption, persistUserRequestActivity, failedActivity)
	e.workflow.AddTransitionActivityWithOptions(persistUserRequestActivity, sgCompleted, e.persistExerciseGeneratedResponse, activityOption, publishActivity)
	e.workflow.AddTransitionActivityWithOptions(publishActivity, sgNoNeed, e.publish, nil, completeActivity)

	e.workflow.RegisterWorkflow(workflowName, e.generateExercise)
}

func (e *exerciseWorkflow) GenerateExercise(ctx context.Context, request *requests.GenerateExerciseRequest) (response *responses.GenerateExerciseResponse, customErr *apperror.CustomError) {
	if request == nil {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, `generate exercise request is nil`, fmt.Errorf("generate exercise request is nil"))
	}
	workflowID := fmt.Sprintf("%s-%s", workflowName, request.IdempotencyKey)

	startWorkflowOptions := temporal_client.StartWorkflowOptions{
		WorkflowID: workflowID,
	}

	workflowRun, err := e.workflow.StartWorkflow(ctx, startWorkflowOptions, e.generateExercise, request)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with workflow `, err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, generateExerciseClientWaitTimeout)
	defer cancel()

	executionData, err := wfQuery(waitCtx, e.temporalClient, queryGenerateExerciseResponse, workflowRun.GetID(), workflowRun.GetRunID(), &ExecutionData{})
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with get result`, err)
	}
	executionDataResult, ok := executionData.(*ExecutionData)
	if !ok || !exerciseResponseReady(executionDataResult) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, `there is wrong with convert result`, fmt.Errorf("generated exercise is missing"))
	}
	return executionDataResult.GenerateExercise, nil
}

func NewExerciseWorkflow(
	workflow temporal_client.WorkflowExecution,
	temporal temporal_client.Temporal,
	exerciseRepository repositories.ExerciseRepository,
	userRepository repositories.UserRepository,
	exerciseCache caches.ExerciseCache,
	deepSeekClient deepseek_client.DeepSeek,
) ExerciseWorkflow {
	exWorkflow := &exerciseWorkflow{
		workflow:           workflow,
		temporalClient:     temporal,
		exerciseRepository: exerciseRepository,
		userRepository:     userRepository,
		exerciseCache:      exerciseCache,
		deepSeekClient:     deepSeekClient,
	}
	exWorkflow.registerActivitiesAndWorkflow()
	return exWorkflow
}
