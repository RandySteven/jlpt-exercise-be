package exercise_workflow

import (
	"testing"

	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

func TestExerciseResponseReadyOnProcessing(t *testing.T) {
	if exerciseResponseReady(nil) {
		t.Fatal("nil execution data was ready")
	}
	processing := &ExecutionData{
		GenerateExercise: &responses.GenerateExerciseResponse{Status: exerciseStatusProcessing},
	}
	if !exerciseResponseReady(processing) {
		t.Fatal("processing response was not ready")
	}
}

func TestExecutionDataUnmarshalIgnoresSignalTrigger(t *testing.T) {
	data := &ExecutionData{
		WorkflowID:     "exercise_workflow-key",
		IdempotencyKey: "key",
		UserRequest:    &requests.GenerateExerciseRequest{IdempotencyKey: "key", JLPTNLevel: 3},
		Prompt:         "prompt",
	}

	if err := data.Unmarshal([]byte("true")); err != nil {
		t.Fatal(err)
	}
	if data.WorkflowID != "exercise_workflow-key" || data.Prompt != "prompt" || data.UserRequest == nil {
		t.Fatalf("trigger signal replaced execution data: %#v", data)
	}
}
