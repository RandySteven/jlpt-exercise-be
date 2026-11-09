package class_workflow

import (
	"encoding/json"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

type NotificationContent struct {
	UserEmail           string `json:"user_email"`
	UserPhoneNumber     string `json:"user_phone_number"`
	ClassName           string `json:"class_name"`
	MidtransRedirectURL string `json:"midtrans_redirect_url"`
}

type ExecutionData struct {
	WorkflowID   string `json:"workflow_id"`
	RunID        string `json:"run_id"`
	CurrActivity string `json:"curr_activity"`
	Status       string `json:"status"`

	Request  *requests.AssignSenseiClassRequest   `json:"request"`
	Response *responses.AssignSenseiClassResponse `json:"response"`

	Class   *models.Class `json:"class"`
	Student *models.User  `json:"student"`

	Onboarding           *models.Onboarding     `json:"onboarding"`
	AllowedUsers         []uint64               `json:"allowed_users"`
	NotificationContents []*NotificationContent `json:"notification_contents"`
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
