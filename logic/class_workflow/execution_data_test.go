package class_workflow

import (
	"errors"
	"testing"

	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

func TestExecutionDataSignalKeepsClassState(t *testing.T) {
	original := &ExecutionData{
		Status:              classStatusProcessing,
		RequestStudentIndex: 0,
		Request:             &requests.AssignSenseiClassRequest{ReferenceID: "class-1", StudentIDs: []uint64{4, 8}},
		Class:               &models.Class{ID: 3, Name: "N3"},
		Response:            &responses.AssignSenseiClassResponse{ClassID: 3, Status: classStatusProcessing},
	}

	if err := original.Unmarshal([]byte("true")); err != nil || original.Class.ID != 3 || original.Request.ReferenceID != "class-1" {
		t.Fatalf("create signal changed state: %#v err %v", original, err)
	}
	if err := original.Unmarshal([]byte(`{"request_student_index":1}`)); err != nil || original.RequestStudentIndex != 1 || original.Class.Name != "N3" || original.Response.ClassID != 3 {
		t.Fatalf("student signal = index %d class %#v err %v", original.RequestStudentIndex, original.Class, err)
	}
}

func TestAssignmentCompletesWhenValidatedStudentsAreNotified(t *testing.T) {
	execData := &ExecutionData{
		Request:  &requests.AssignSenseiClassRequest{StudentIDs: []uint64{4, 8, 9}},
		Response: &responses.AssignSenseiClassResponse{Status: classStatusProcessing},
	}

	execData.recordStudent(&responses.StudentResponse{StudentID: 4, Status: studentStatusRejected})
	execData.continueOrFinish()
	if execData.assignmentComplete() || execData.CurrActivity != validateUserActivityName {
		t.Fatalf("rejected student finished the workflow: complete %v activity %s", execData.assignmentComplete(), execData.CurrActivity)
	}

	execData.recordStudent(&responses.StudentResponse{StudentID: 8, Status: studentStatusNotified})
	execData.continueOrFinish()
	if execData.assignmentComplete() || execData.Response.Status != classStatusProcessing || execData.Response.SuccessCount != 1 {
		t.Fatalf("workflow completed before the last validated student: %#v", execData.Response)
	}

	execData.recordStudent(&responses.StudentResponse{StudentID: 9, Status: studentStatusNotified})
	execData.continueOrFinish()
	if !execData.assignmentComplete() || execData.CurrActivity != finishWorkflowActivityName || execData.Response.Status != classStatusCompleted || execData.Response.SuccessCount != 2 {
		t.Fatalf("workflow did not complete after midtrans notifications: %#v activity %s", execData.Response, execData.CurrActivity)
	}
}

func TestPushNotificationDoesNotFinishUntilSendSucceeds(t *testing.T) {
	request := &requests.AssignSenseiClassRequest{StudentIDs: []uint64{4, 8}}
	execData := &ExecutionData{
		Request:             request,
		RequestStudentIndex: 1,
		Response:            &responses.AssignSenseiClassResponse{Status: classStatusProcessing},
		Student:             &models.User{ID: 8},
		Onboarding:          &models.Onboarding{Name: "Aya", Email: "aya@example.com"},
		NotificationContents: &NotificationContent{
			UserEmail:           "aya@example.com",
			ClassName:           "N3",
			MidtransRedirectURL: "https://pay.example/snap",
		},
	}
	execData.recordStudent(&responses.StudentResponse{StudentID: 4, Status: studentStatusRejected})

	smtp := &stubSMTP{err: errors.New("smtp down")}
	workflow := &classWorkflow{smtpClient: smtp}
	if _, err := workflow.pushNotification(t.Context(), execData); err == nil || execData.studentFinished(8) || execData.assignmentComplete() {
		t.Fatalf("failed notification was recorded: finished %v complete %v err %v", execData.studentFinished(8), execData.assignmentComplete(), err)
	}
	if len(smtp.sent) != 0 {
		t.Fatalf("email sent on failure: %#v", smtp.sent)
	}

	smtp.err = nil
	if _, err := workflow.pushNotification(t.Context(), execData); err != nil {
		t.Fatal(err)
	}
	if !execData.assignmentComplete() || execData.CurrActivity != finishWorkflowActivityName || smtp.sent[0] != "aya@example.com" {
		t.Fatalf("notification did not complete the workflow: activity %s sent %#v response %#v", execData.CurrActivity, smtp.sent, execData.Response)
	}

	if _, err := workflow.finishWorkflow(t.Context(), execData); err != nil || execData.Response.Status != classStatusCompleted || execData.CurrActivity != "" {
		t.Fatalf("finish = status %s activity %q err %v", execData.Response.Status, execData.CurrActivity, err)
	}
}

type stubSMTP struct {
	err  error
	sent []string
}

func (s *stubSMTP) SendEmail(email string, subject string, body string) error {
	if s.err != nil {
		return s.err
	}
	s.sent = append(s.sent, email)
	return nil
}

func (s *stubSMTP) SendHTMLTemplate(email string, subject string, templateName string, data any) error {
	return s.err
}
