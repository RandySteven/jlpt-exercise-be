package class_workflow

import (
	"context"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/entities/models"
)

func (c *classWorkflow) persistClass(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	class := &models.Class{
		SenseiID:  execData.Request.SenseiID,
		Name:      execData.Request.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	class, err := c.classRepository.Save(ctx, class)
	if err != nil {
		return nil, err
	}
	execData.Class = class

	execData.SetActivity(validateUserActivityName)

	return execData, nil
}
