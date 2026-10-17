package onboarding_workflow

import (
	"bytes"
	"encoding/json"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

type ExecutionData struct {
	Request         *requests.RegisterRequest   `json:"request"`
	Response        *responses.RegisterResponse `json:"response"`
	WorkflowID      string                      `json:"workflow_id"`
	RunID           string                      `json:"run_id"`
	Status          string                      `json:"status"`
	Onboarding      *models.Onboarding          `json:"onboarding"`
	User            *models.User                `json:"user"`
	ActivationToken string                      `json:"activation_token"`
	CurrActivity    string                      `json:"curr_activity"`
	NextActivity    string                      `json:"next_activity"`
}

func (e *ExecutionData) GetActivity() string {
	return e.CurrActivity
}

func (e *ExecutionData) SetActivity(activityName string) {
	e.CurrActivity = activityName
}

func (e *ExecutionData) Marshal() ([]byte, error) {
	return json.Marshal(e)
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

	var asRequest requests.RegisterRequest
	if err := json.Unmarshal(trimmed, &asRequest); err == nil && isRegisterRequestPatch(&asRequest) {
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

func isRegisterRequestPatch(req *requests.RegisterRequest) bool {
	return req.Email != "" || req.Username != "" || req.FirstName != "" || req.PhoneNumber != ""
}

var _ temporal_client.ExecutionData = (*ExecutionData)(nil)
