package payment_workflow

import (
	"fmt"

	temporal_client "github.com/RandySteven/go-cook/temporal"
)

func setQueryWorkflowResponse(execData temporal_client.ExecutionData) error {
	current, ok := execData.(*ExecutionData)
	if !ok {
		return fmt.Errorf("response is not ready")
	}
	if !queryWorkflowResponseReady(current) {
		return fmt.Errorf("response is not ready")
	}
	return nil
}

func queryWorkflowResponseReady(current *ExecutionData) bool {
	return current != nil && current.Response != nil && current.Response.SubscriptionID != 0
}
