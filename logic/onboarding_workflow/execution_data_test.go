package onboarding_workflow

import (
	"testing"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

func TestExecutionDataRoundTrip(t *testing.T) {
	original := &ExecutionData{
		WorkflowID:   "wf",
		RunID:        "run",
		CurrActivity: persistOnboardingRequestActivity,
		Request:      &requests.RegisterRequest{Email: "ada@example.com", Username: "ada"},
		Response:     &responses.RegisterResponse{UserID: 4, ActivationToken: "token"},
	}
	original.SetActivity(registerUserActivity)
	if original.GetActivity() != registerUserActivity {
		t.Fatalf("activity = %s", original.GetActivity())
	}

	encoded, err := original.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	decoded := &ExecutionData{}
	if err = decoded.Unmarshal(encoded); err != nil {
		t.Fatal(err)
	}
	if decoded.WorkflowID != "wf" || decoded.Request.Email != "ada@example.com" || decoded.Response.ActivationToken != "token" || decoded.CurrActivity != registerUserActivity {
		t.Fatalf("decoded = %#v", decoded)
	}

	var nilData *ExecutionData
	if nilData.Clone().(*ExecutionData) != nil {
		t.Fatal("nil execution data did not clone to a nil pointer")
	}
	cloned := original.Clone().(*ExecutionData)
	cloned.Request.Email = "changed@example.com"
	if original.Request.Email != "ada@example.com" {
		t.Fatal("clone shared the request")
	}
}

func TestExecutionDataUnmarshalSpecialPayloads(t *testing.T) {
	data := &ExecutionData{WorkflowID: "wf", RunID: "run"}
	for _, payload := range []string{"", "   ", "null", "true", "false"} {
		if err := data.Unmarshal([]byte(payload)); err != nil || data.WorkflowID != "wf" {
			t.Fatalf("payload %q replaced execution data: %#v err %v", payload, data, err)
		}
	}

	if err := data.Unmarshal([]byte(`{"email":"ada@example.com","username":"ada"}`)); err != nil {
		t.Fatal(err)
	}
	if data.WorkflowID != "wf" || data.Request == nil || data.Request.Email != "ada@example.com" || data.Request.Username != "ada" {
		t.Fatalf("patch = %#v", data)
	}

	if err := data.Unmarshal([]byte(`{"not":`)); err == nil {
		t.Fatal("invalid json was accepted")
	}
}

func TestExecutionDataSatisfiesInterface(t *testing.T) {
	var _ temporal_client.ExecutionData = (*ExecutionData)(nil)
}
