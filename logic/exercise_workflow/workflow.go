package exercise_workflow

import (
	temporal_client "github.com/RandySteven/go-cook/temporal"
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
)

type (
	exerciseWorkflow struct {
		workflow       temporal_client.WorkflowExecution
		temporalClient temporal_client.Temporal
	}

	ExerciseWorkflow interface {
		GenerateExercise()
	}
)

func (e *exerciseWorkflow) registerActivitiesAndWorkflow() {
	e.workflow.AddTransitionActivityWithOptions(persistUserRequestActivity, sgPersist, e.persistUserRequest, nil, readPromptTxtActivity)
	e.workflow.AddTransitionActivityWithOptions(readPromptTxtActivity, sgNoNeed, e.readPromptTxt, nil, generateExerciseActivity, failedActivity)
	e.workflow.AddTransitionActivityWithOptions(generateExerciseActivity, sgProcessing, e.generate, nil, saveExerciseActivity, failedActivity)

	e.workflow.AddTransitionActivityWithOptions(saveExerciseActivity, sgCompleted, e.saveExercise, nil, publishActivity, failedActivity)
	e.workflow.AddTransitionActivityWithOptions(publishActivity, sgNoNeed, e.publish, nil, completeActivity)

	e.workflow.RegisterWorkflow(workflowName, e.generateExercise)
}