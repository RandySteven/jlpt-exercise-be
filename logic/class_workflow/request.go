package class_workflow

import (
	"context"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/entities/models"
)

func (c *classWorkflow) persistClass(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	class := &models.Class{
		SenseiID:    execData.Request.SenseiID,
		Name:        execData.Request.Name,
		ReferenceID: execData.Request.ReferenceID,
		Schedules:   classSchedules(execData),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	class, err := c.classRepository.Save(ctx, class)
	if err != nil {
		return nil, err
	}
	execData.Class = class
	if execData.Response != nil {
		execData.Response.ClassID = class.ID
	}

	execData.SetActivity(validateUserActivityName)
	return execData, nil
}

func classSchedules(execData *ExecutionData) []models.ClassSchedule {
	if execData == nil || execData.Request == nil {
		return nil
	}
	schedules := make([]models.ClassSchedule, 0, len(execData.Request.Schedules))
	for _, schedule := range execData.Request.Schedules {
		schedules = append(schedules, models.ClassSchedule{
			Day:       schedule.Day,
			StartTime: schedule.StartTime,
			EndTime:   schedule.EndTime,
		})
	}
	return schedules
}
