package payment_workflow

import (
	"encoding/json"

	temporal_client "github.com/RandySteven/go-cook/temporal"
)

type ExecutionData struct {
	WorkflowID   string `json:"workflow_id"`
	RunID        string `json:"run_id"`
	CurrActivity string `json:"curr_activity"`
	Status       string `json:"status"`
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
	return json.Unmarshal(data, e)
}
