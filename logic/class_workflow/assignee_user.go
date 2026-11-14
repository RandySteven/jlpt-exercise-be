package class_workflow

import (
	"context"
	"database/sql"
	"errors"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/enums"
)

func (c *classWorkflow) validateUser(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData.Request == nil || len(execData.Request.StudentIDs) != 1 {
		return nil, errors.New("exactly one student is required")
	}

	userID := execData.Request.StudentIDs[execData.RequestStudentIndex]
	user, err := c.userRepository.FindByID(ctx, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if user == nil || user.Status != `ACTIVE` {
		execData.SetActivity(rejectStudentActivityName)
		return execData, nil
	}

	execData.Student = user
	execData.SetActivity(assignUserToClassActivityName)
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

	execData.SetActivity(publishClassActivityName)

	return execData, nil
}
