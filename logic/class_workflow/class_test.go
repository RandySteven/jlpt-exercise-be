package class_workflow

import (
	"errors"
	"testing"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"go.temporal.io/sdk/testsuite"
	temporalworkflow "go.temporal.io/sdk/workflow"
)

func TestAssignSenseiClassWorkflowRejectsIncompleteRequest(t *testing.T) {
	workflow := &classWorkflow{}
	if _, err := workflow.AssignSenseiClassWorkflow(t.Context(), nil, 0); err == nil || err.ErrType != apperror.ErrBadRequest {
		t.Fatalf("nil request err = %#v", err)
	}
	request := &requests.AssignSenseiClassRequest{ReferenceID: "class-1", StudentIDs: []uint64{4}}
	if _, err := workflow.AssignSenseiClassWorkflow(t.Context(), request, 1); err == nil || err.ErrType != apperror.ErrBadRequest {
		t.Fatalf("index err = %#v", err)
	}
}

func TestCreateClassReturnsAfterEveryValidatedStudentIsNotified(t *testing.T) {
	originalInit := wfInit
	originalExecute := wfExecute
	t.Cleanup(func() {
		wfInit = originalInit
		wfExecute = originalExecute
	})

	wfInit = func(_ temporalworkflow.Context, _ temporal_client.WorkflowExecution, queryType string, data temporal_client.ExecutionData, _ func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error) {
		if queryType != queryAssignSenseiClassResponse {
			return nil, errors.New("unexpected query")
		}
		return data, nil
	}
	wfExecute = func(_ temporalworkflow.Context, _ temporal_client.WorkflowExecution, data temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
		return data, nil
	}

	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow((&classWorkflow{}).createClass, &requests.AssignSenseiClassRequest{ReferenceID: "class-1", StudentIDs: []uint64{4}})
	if !env.IsWorkflowCompleted() || env.GetWorkflowError() == nil {
		t.Fatal("workflow completed before a midtrans notification")
	}

	wfExecute = func(_ temporalworkflow.Context, _ temporal_client.WorkflowExecution, data temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
		current := data.(*ExecutionData)
		current.recordStudent(&responses.StudentResponse{StudentID: current.Request.StudentIDs[0], Status: studentStatusNotified})
		current.Response.Status = classStatusCompleted
		return data, nil
	}
	env = suite.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow((&classWorkflow{}).createClass, &requests.AssignSenseiClassRequest{ReferenceID: "class-1", StudentIDs: []uint64{4}})
	if !env.IsWorkflowCompleted() || env.GetWorkflowError() != nil {
		t.Fatalf("workflow error: %v", env.GetWorkflowError())
	}
	var response *responses.AssignSenseiClassResponse
	if err := env.GetWorkflowResult(&response); err != nil {
		t.Fatal(err)
	}
	if response.Status != classStatusCompleted || len(response.Students) != 1 || response.Students[0].Status != studentStatusNotified {
		t.Fatalf("response = %#v", response)
	}
}
