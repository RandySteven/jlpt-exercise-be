package class_workflow

import (
	"bytes"
	"encoding/json"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

const (
	classStatusProcessing = "PROCESSING"
	classStatusCompleted  = "COMPLETED"

	studentStatusRejected = "REJECTED"
	studentStatusNotified = "NOTIFIED"
)

type NotificationContent struct {
	UserEmail           string `json:"user_email"`
	UserPhoneNumber     string `json:"user_phone_number"`
	ClassName           string `json:"class_name"`
	MidtransRedirectURL string `json:"midtrans_redirect_url"`
}

type addStudentSignal struct {
	RequestStudentIndex uint64 `json:"request_student_index"`
}

type ExecutionData struct {
	WorkflowID   string `json:"workflow_id"`
	RunID        string `json:"run_id"`
	CurrActivity string `json:"curr_activity"`
	Status       string `json:"status"`

	Request             *requests.AssignSenseiClassRequest   `json:"request"`
	Response            *responses.AssignSenseiClassResponse `json:"response"`
	RequestStudentIndex uint64                               `json:"request_student_index"`

	Class   *models.Class   `json:"class"`
	Student *models.User    `json:"student"`
	Payment *models.Payment `json:"payment"`

	Onboarding           *models.Onboarding   `json:"onboarding"`
	NotificationContents *NotificationContent `json:"notification_contents"`
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

	var flag bool
	if err := json.Unmarshal(trimmed, &flag); err == nil {
		return nil
	}

	if index, ok := studentIndexSignal(trimmed); ok {
		e.RequestStudentIndex = index
		return nil
	}

	var next ExecutionData
	if err := json.Unmarshal(trimmed, &next); err != nil {
		return err
	}
	*e = next
	return nil
}

func studentIndexSignal(data []byte) (uint64, bool) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return 0, false
	}
	if len(raw) != 1 {
		return 0, false
	}
	value, ok := raw["request_student_index"]
	if !ok {
		return 0, false
	}
	var index uint64
	if err := json.Unmarshal(value, &index); err != nil {
		return 0, false
	}
	return index, true
}

func (e *ExecutionData) studentStatus(studentID uint64) (string, bool) {
	if e == nil || e.Response == nil {
		return "", false
	}
	for _, student := range e.Response.Students {
		if student != nil && student.StudentID == studentID {
			return student.Status, true
		}
	}
	return "", false
}

func (e *ExecutionData) studentFinished(studentID uint64) bool {
	status, ok := e.studentStatus(studentID)
	return ok && (status == studentStatusRejected || status == studentStatusNotified)
}

func (e *ExecutionData) recordStudent(student *responses.StudentResponse) {
	if e == nil || student == nil {
		return
	}
	if e.Response == nil {
		e.Response = &responses.AssignSenseiClassResponse{}
	}
	for i, existing := range e.Response.Students {
		if existing != nil && existing.StudentID == student.StudentID {
			e.Response.Students[i] = student
			e.Response.SuccessCount = e.notifiedCount()
			return
		}
	}
	e.Response.Students = append(e.Response.Students, student)
	e.Response.SuccessCount = e.notifiedCount()
}

func (e *ExecutionData) notifiedCount() uint64 {
	if e == nil || e.Response == nil {
		return 0
	}
	var count uint64
	for _, student := range e.Response.Students {
		if student != nil && student.Status == studentStatusNotified {
			count++
		}
	}
	return count
}

// assignmentComplete is true once every requested student has an outcome and
// every validated student has been sent the Midtrans notification.
func (e *ExecutionData) assignmentComplete() bool {
	if e == nil || e.Request == nil || len(e.Request.StudentIDs) == 0 {
		return false
	}
	for _, studentID := range e.Request.StudentIDs {
		if !e.studentFinished(studentID) {
			return false
		}
	}
	return e.notifiedCount() == e.validatedCount()
}

func (e *ExecutionData) validatedCount() uint64 {
	if e == nil || e.Request == nil || e.Response == nil {
		return 0
	}
	var count uint64
	for _, studentID := range e.Request.StudentIDs {
		status, ok := e.studentStatus(studentID)
		if ok && status != studentStatusRejected {
			count++
		}
	}
	return count
}

func (e *ExecutionData) continueOrFinish() {
	if e == nil {
		return
	}
	if e.assignmentComplete() {
		e.Status = classStatusCompleted
		if e.Response != nil {
			e.Response.Status = classStatusCompleted
		}
		e.SetActivity(finishWorkflowActivityName)
		return
	}
	e.Status = classStatusProcessing
	if e.Response != nil {
		e.Response.Status = classStatusProcessing
	}
	e.SetActivity(validateUserActivityName)
}
