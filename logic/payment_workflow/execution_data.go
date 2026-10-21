package payment_workflow

import (
	"encoding/json"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
)

type ExecutionData struct {
	WorkflowID   string `json:"workflow_id"`
	RunID        string `json:"run_id"`
	CurrActivity string `json:"curr_activity"`
	Status       string `json:"status"`

	MidtransTransaction *midtrans_client.TransactionResponse `json:"midtrans_transaction"`
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
