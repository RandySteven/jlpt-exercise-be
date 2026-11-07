package class_workflow

import (
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"go.temporal.io/sdk/workflow"
)

func (c *classWorkflow) createClass(ctx workflow.Context, request *requests.AssignSenseiClassRequest) (*responses.AssignSenseiClassResponse, error) {
	return nil, nil
}
