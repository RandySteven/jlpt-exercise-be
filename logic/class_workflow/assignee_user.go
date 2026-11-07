package class_workflow

import (
	"context"
	"database/sql"
	"errors"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
)

func (c *classWorkflow) validateUser(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	for _, userID := range execData.Request.StudentIDs {
		user, err := c.userRepository.FindByID(ctx, userID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if user == nil || user.Status != `ACTIVE` {
			continue
		}

		execData.AllowedUsers = append(execData.AllowedUsers, userID)
	}
	return execData, nil
}

func (c *classWorkflow) assignUserToClass(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	class, err := c.classRepository.FindByID(ctx, execData.Class.ID)
	if err != nil {
		return nil, err
	}
	if class == nil {
		return nil, errors.New("class not found")
	}

	customErr := c.transaction.RunInTx(ctx, func(ctx context.Context) *apperror.CustomError {
		for _, userID := range execData.AllowedUsers {
			_, err := c.classUserRepository.Save(ctx, &models.ClassUser{
				ClassID: class.ID,
				UserID:  userID,
			})
			if err != nil {
				return apperror.DatabaseErrorErr()
			}
		}
		return nil
	})
	if customErr != nil {
		return nil, customErr
	}

	

	return execData, nil
}
