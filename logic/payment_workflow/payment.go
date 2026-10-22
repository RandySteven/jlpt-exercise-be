package payment_workflow

import (
	"fmt"

	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"go.temporal.io/sdk/workflow"
)

func (p *paymentWorkflow) payment(ctx workflow.Context, request *requests.CreateSubscriptionRequest, userID uint64) (*responses.CreateSubscriptionResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("subscription request is nil")
	}
	if userID == 0 {
		return nil, fmt.Errorf("user id is missing")
	}

	info := workflow.GetInfo(ctx)
	execData := &ExecutionData{
		CurrActivity: checkPlanActivityName,
		WorkflowID:   info.WorkflowExecution.ID,
		RunID:        info.WorkflowExecution.RunID,
		Request:      request,
		UserID:       userID,
		Status:       subscriptionStatusPending,
	}

	executionDataResult, err := wfInit(ctx, p.workflow, queryPaymentWorkflow, execData, setQueryWorkflowResponse)
	if err != nil {
		return nil, err
	}

	executionData, ok := executionDataResult.(*ExecutionData)
	if !ok {
		return nil, fmt.Errorf("execution data is not ready")
	}

	if _, err = wfExecute(ctx, p.workflow, executionData); err != nil {
		return nil, err
	}
	if !queryWorkflowResponseReady(executionData) {
		return nil, fmt.Errorf("subscription response is missing")
	}

	return executionData.Response, nil
}
