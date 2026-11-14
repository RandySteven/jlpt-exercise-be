package class_workflow

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
	if current == nil || current.Response == nil || current.Request == nil {
		return false
	}
	if current.RequestStudentIndex >= uint64(len(current.Request.StudentIDs)) {
		return false
	}
	studentID := current.Request.StudentIDs[current.RequestStudentIndex]
	return current.studentFinished(studentID)
}
