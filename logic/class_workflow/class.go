package class_workflow

import (
	"context"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"go.temporal.io/sdk/workflow"
)

func (c *classWorkflow) createClass(ctx workflow.Context, request *requests.AssignSenseiClassRequest, index uint64) (*responses.AssignSenseiClassResponse, error) {
	class, err := c.classRepository.FindByReferenceID(context.Background(), request.ReferenceID)
	if err != nil {
		return nil, err
	}

	if class == nil {
		executionData := &ExecutionData{}
		_, err := wfInit(ctx, c.workflow, queryAssignSenseiClassResponse, executionData, func(executionData temporal_client.ExecutionData) error {
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	//wfExec

	return nil, nil
}
