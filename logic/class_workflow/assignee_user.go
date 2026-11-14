package class_workflow

import (
	"context"
	"database/sql"
	"errors"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
)

func (c *classWorkflow) validateUser(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Request == nil || len(execData.Request.StudentIDs) == 0 {
		return nil, errors.New("students are required")
	}
	if execData.RequestStudentIndex >= uint64(len(execData.Request.StudentIDs)) {
		return nil, errors.New("student index is out of range")
	}

	userID := execData.Request.StudentIDs[execData.RequestStudentIndex]
	if execData.studentFinished(userID) {
		execData.continueOrFinish()
		return execData, nil
	}

	execData.Student = nil
	execData.Onboarding = nil
	execData.Payment = nil
	execData.NotificationContents = nil

	user, err := c.userRepository.FindByID(ctx, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if user == nil || user.Status != `ACTIVE` {
		execData.SetActivity(rejectStudentActivityName)
		return execData, nil
	}
	if user.OnboardingID == 0 || c.onboardingRepository == nil {
		return nil, errors.New("student onboarding is missing")
	}

	onboarding, err := c.onboardingRepository.FindByID(ctx, user.OnboardingID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if onboarding == nil || onboarding.Email == "" {
		return nil, errors.New("student onboarding is missing")
	}

	execData.Student = user
	execData.Onboarding = onboarding
	execData.SetActivity(assignUserToClassActivityName)
	return execData, nil
}

func (c *classWorkflow) rejectStudent(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || execData.Request == nil || execData.RequestStudentIndex >= uint64(len(execData.Request.StudentIDs)) {
		return nil, errors.New("student is missing")
	}

	execData.recordStudent(&responses.StudentResponse{
		StudentID: execData.Request.StudentIDs[execData.RequestStudentIndex],
		Status:    studentStatusRejected,
	})
	execData.continueOrFinish()
	return execData, nil
}

func (c *classWorkflow) assignUserToClass(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData.Class == nil {
		return nil, errors.New("class not found")
	}
	if execData.Student == nil {
		return nil, errors.New("student not found")
	}

	class, err := c.classRepository.FindByID(ctx, execData.Class.ID)
	if err != nil {
		return nil, err
	}
	if class == nil {
		return nil, errors.New("class not found")
	}

	customErr := c.transaction.RunInTx(ctx, func(ctx context.Context) *apperror.CustomError {
		_, err := c.classUserRepository.Save(ctx, &models.ClassUser{
			ClassID: class.ID,
			UserID:  execData.Student.ID,
			Status:  enums.ClassUserStatusPending.String(),
		})
		if err != nil {
			return apperror.DatabaseErrorErr()
		}
		return nil
	})
	if customErr != nil {
		return nil, customErr
	}

	execData.SetActivity(initiatedPaymentForStudentActivityName)
	return execData, nil
}
