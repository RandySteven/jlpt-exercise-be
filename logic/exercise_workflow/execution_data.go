package exercise_workflow

import (
	"bytes"
	"encoding/json"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

type ExecutionData struct {
	WorkflowID   string `json:"workflow_id"`
	RunID        string `json:"run_id"`
	CurrActivity string `json:"curr_activity"`
	Status       string `json:"status"`
	Prompt       string `json:"prompt"`
	PromptResult string `json:"prompt_result"`

	IdempotencyKey   string                              `json:"idempotency_key"`
	UserID           uint64                              `json:"user_id"`
	UserRequest      *requests.GenerateExerciseRequest   `json:"user_request"`
	GenerateExercise *responses.GenerateExerciseResponse `json:"generate_exercise"`

	Exercise *models.Exercise `json:"exercise"`
	User     *models.User     `json:"user"`
}

var _ temporal_client.ExecutionData = (*ExecutionData)(nil)

func (e *ExecutionData) Clone() temporal_client.ExecutionData {
	c := *e
	return &c
}

func (e *ExecutionData) GetActivity() string {
	return e.CurrActivity
}

func (e *ExecutionData) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ExecutionData) SetActivity(activityName string) {
	e.CurrActivity = activityName
}

func (e *ExecutionData) Unmarshal(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil
	}

	var trigger bool
	if err := json.Unmarshal(trimmed, &trigger); err == nil {
		return nil
	}

	var next ExecutionData
	if err := json.Unmarshal(trimmed, &next); err != nil {
		return err
	}
	*e = next
	return nil
}
