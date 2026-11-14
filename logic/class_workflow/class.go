package class_workflow

import (
	"context"
	"fmt"

	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"go.temporal.io/sdk/workflow"
)

func (c *classWorkflow) createClass(ctx workflow.Context, request *requests.AssignSenseiClassRequest) (*responses.AssignSenseiClassResponse, error) {
	if request == nil || request.ReferenceID == "" || len(request.StudentIDs) == 0 {
		return nil, fmt.Errorf("assign sensei class request is incomplete")
	}

	info := workflow.GetInfo(ctx)
	now := workflow.Now(ctx)
	executionData := &ExecutionData{
		WorkflowID:   info.WorkflowExecution.ID,
		RunID:        info.WorkflowExecution.RunID,
		CurrActivity: persistClassActivityName,
		Status:       classStatusProcessing,
		Request:      request,
		Response: &responses.AssignSenseiClassResponse{
			Status:    classStatusProcessing,
			CreatedAt: now,
			Students:  []*responses.StudentResponse{},
		},
	}

	executionDataResult, err := wfInit(ctx, c.workflow, queryAssignSenseiClassResponse, executionData, setQueryWorkflowResponse)
	if err != nil {
		return nil, err
	}
	executionData, ok := executionDataResult.(*ExecutionData)
	if !ok {
		return nil, fmt.Errorf("execution data is not ready")
	}

	if _, err = wfExecute(ctx, c.workflow, executionData); err != nil {
		return nil, err
	}
	if !executionData.assignmentComplete() {
		return nil, fmt.Errorf("validated students are still waiting for a midtrans notification")
	}
	return executionData.Response, nil
}

func (c *classWorkflow) finishWorkflow(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData == nil || !execData.assignmentComplete() {
		return nil, fmt.Errorf("validated students are still waiting for a midtrans notification")
	}
	execData.Status = classStatusCompleted
	if execData.Response != nil {
		execData.Response.Status = classStatusCompleted
		if execData.Class != nil {
			execData.Response.ClassID = execData.Class.ID
		}
	}
	execData.SetActivity("")
	return execData, nil
}
