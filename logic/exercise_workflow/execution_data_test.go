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

func TestExecutionDataRoundTrip(t *testing.T) {
	original := &ExecutionData{
		WorkflowID:     "exercise_workflow-key",
		RunID:          "run",
		IdempotencyKey: "key",
		Prompt:         "prompt",
	}
	original.SetActivity(generateExerciseActivity)
	if original.GetActivity() != generateExerciseActivity {
		t.Fatalf("activity = %s", original.GetActivity())
	}

	encoded, err := original.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	decoded := &ExecutionData{WorkflowID: "stale"}
	if err = decoded.Unmarshal(encoded); err != nil {
		t.Fatal(err)
	}
	if decoded.WorkflowID != original.WorkflowID || decoded.Prompt != "prompt" || decoded.CurrActivity != generateExerciseActivity {
		t.Fatalf("decoded = %#v", decoded)
	}

	cloned := original.Clone().(*ExecutionData)
	cloned.Prompt = "changed"
	if original.Prompt != "prompt" {
		t.Fatal("clone shared the prompt string holder unexpectedly")
	}

	for _, payload := range []string{"", "   ", "null", "false"} {
		kept := &ExecutionData{WorkflowID: "keep"}
		if err = kept.Unmarshal([]byte(payload)); err != nil || kept.WorkflowID != "keep" {
			t.Fatalf("payload %q replaced execution data: %#v err %v", payload, kept, err)
		}
	}
	if err = decoded.Unmarshal([]byte(`{"workflow_id":`)); err == nil {
		t.Fatal("invalid json was accepted")
	}
}

func TestExerciseResponseReadyWhenExerciseExists(t *testing.T) {
	if exerciseResponseReady(&ExecutionData{}) {
		t.Fatal("missing response was ready")
	}
	completed := &ExecutionData{
		GenerateExercise: &responses.GenerateExerciseResponse{
			Exercise: map[string][]*responses.ExerciseQuestion{
				"grammar": {{Question: "q1"}},
			},
		},
	}
	if !exerciseResponseReady(completed) {
		t.Fatal("response with exercise content was not ready")
	}
	if exerciseResponseReady(&ExecutionData{GenerateExercise: &responses.GenerateExerciseResponse{}}) {
		t.Fatal("empty response was ready")
	}
}
