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

	//sg event
	sgProcessing = `sg_processing_generation`
	sgCompleted  = `sg_completed_generation`
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
