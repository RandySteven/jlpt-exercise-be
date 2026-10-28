package payment_workflow

import (
	"bytes"
	"encoding/json"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
)

type ExecutionData struct {
	WorkflowID   string `json:"workflow_id"`
	RunID        string `json:"run_id"`
	CurrActivity string `json:"curr_activity"`
	Status       string `json:"status"`

	UserID         uint64                               `json:"user_id"`
	CustomerEmail  string                               `json:"customer_email"`
	CustomerName   string                               `json:"customer_name"`
	Request        *requests.CreateSubscriptionRequest  `json:"request"`
	ConfirmRequest *requests.PaymentConfirmationRequest `json:"confirm_request"`
	MidtransStatus string                               `json:"midtrans_status"`

	Plan         *models.SubscriptionPlan `json:"plan"`
	Payment      *models.Payment          `json:"payment"`
	Subscription *models.Subscription     `json:"subscription"`

	Response *responses.CreateSubscriptionResponse `json:"response"`

	MidtransTransaction *midtrans_client.TransactionResponse `json:"midtrans_transaction"`
}

var _ temporal_client.ExecutionData = (*ExecutionData)(nil)

func (e *ExecutionData) Clone() temporal_client.ExecutionData {
	if e == nil {
		return (*ExecutionData)(nil)
	}
	data, err := e.Marshal()
	if err != nil {
		cp := *e
		return &cp
	}
	clone := &ExecutionData{}
	if err := json.Unmarshal(data, clone); err != nil {
		cp := *e
		return &cp
	}
	return clone
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

	var flag bool
	if err := json.Unmarshal(trimmed, &flag); err == nil {
		return nil
	}

	if confirm, ok := confirmationSignal(trimmed); ok {
		e.ConfirmRequest = confirm
		return nil
	}

	var asRequest requests.CreateSubscriptionRequest
	if err := json.Unmarshal(trimmed, &asRequest); err == nil && (asRequest.IdempotencyKey != "" || asRequest.PlanID != 0) {
		e.Request = &asRequest
		return nil
	}

	var next ExecutionData
	if err := json.Unmarshal(trimmed, &next); err != nil {
		return err
	}
	*e = next
	return nil
}

func confirmationSignal(data []byte) (*requests.PaymentConfirmationRequest, bool) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, false
	}
	if _, ok := raw["status"]; !ok {
		return nil, false
	}
	for key := range raw {
		switch key {
		case "idempotency_key", "plan_id", "status":
		default:
			return nil, false
		}
	}
	var confirm requests.PaymentConfirmationRequest
	if err := json.Unmarshal(data, &confirm); err != nil {
		return nil, false
	}
	return &confirm, true
}
