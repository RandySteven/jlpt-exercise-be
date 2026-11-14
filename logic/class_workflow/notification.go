package class_workflow

import (
	"context"
	"errors"
	"fmt"

	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

func (c *classWorkflow) pushNotification(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Student == nil || execData.NotificationContents == nil {
		return nil, errors.New("midtrans notification is missing")
	}
	if execData.NotificationContents.UserEmail == "" || execData.NotificationContents.MidtransRedirectURL == "" {
		return nil, errors.New("midtrans notification is missing")
	}
	if c.smtpClient == nil {
		return nil, errors.New("smtp client is missing")
	}

	subject := "Complete your class payment"
	if execData.NotificationContents.ClassName != "" {
		subject = fmt.Sprintf("Complete your payment for %s", execData.NotificationContents.ClassName)
	}
	body := fmt.Sprintf("Pay for your class: %s", execData.NotificationContents.MidtransRedirectURL)
	if err := c.smtpClient.SendEmail(execData.NotificationContents.UserEmail, subject, body); err != nil {
		return nil, err
	}

	name := ""
	if execData.Onboarding != nil {
		name = execData.Onboarding.Name
	}
	execData.recordStudent(&responses.StudentResponse{
		StudentID: execData.Student.ID,
		Name:      name,
		Status:    studentStatusNotified,
	})
	if execData.Class != nil && execData.Response != nil {
		execData.Response.ClassID = execData.Class.ID
	}
	execData.continueOrFinish()
	return execData, nil
}
