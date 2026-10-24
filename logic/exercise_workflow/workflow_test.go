package exercise_workflow

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	temporalmocks "github.com/RandySteven/jlpt-exercise-be/mocks/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func TestNewExerciseWorkflowRegistersPipeline(t *testing.T) {
	execution := newWorkflowDouble(t)
	got := NewExerciseWorkflow(execution.exec, nil, nil, nil, nil, nil)
	if got == nil {
		t.Fatal("workflow was nil")
	}
	if len(execution.workflows) != 1 || execution.workflows[0] != workflowName {
		t.Fatalf("registered workflows = %#v", execution.workflows)
	}

	want := []recordedActivity{
		{name: persistUserRequestActivity, signal: sgNoNeed, next: []string{readPromptTxtActivity, failedActivity}, timeout: 25 * time.Second, attempts: 5, nonRetryable: []string{exerciseInputErrorType}},
		{name: readPromptTxtActivity, signal: sgNoNeed, next: []string{generateExerciseActivity, failedActivity}, timeout: 25 * time.Second, attempts: 5, nonRetryable: []string{exerciseInputErrorType}},
		{name: generateExerciseActivity, signal: sgProcessing, next: []string{persistGeneratedResponseActivity, failedActivity}, timeout: 5 * time.Minute, attempts: 5},
		{name: persistGeneratedResponseActivity, signal: sgCompleted, next: []string{publishActivity}, timeout: 25 * time.Second, attempts: 5, nonRetryable: []string{exerciseInputErrorType}},
		{name: publishActivity, signal: sgNoNeed, timeout: 25 * time.Second, attempts: 5, nonRetryable: []string{exerciseInputErrorType}},
	}
	if len(execution.activities) != len(want) {
		t.Fatalf("registered %d activities, want %d", len(execution.activities), len(want))
	}
	for i, activity := range want {
		gotActivity := execution.activities[i]
		if gotActivity.name != activity.name || gotActivity.signal != activity.signal || gotActivity.timeout != activity.timeout || gotActivity.attempts != activity.attempts {
			t.Fatalf("activity %d = %#v, want %#v", i, gotActivity, activity)
		}
		if strings.Join(gotActivity.next, ",") != strings.Join(activity.next, ",") {
			t.Fatalf("activity %s next = %#v, want %#v", activity.name, gotActivity.next, activity.next)
		}
		if strings.Join(gotActivity.nonRetryable, ",") != strings.Join(activity.nonRetryable, ",") {
			t.Fatalf("activity %s non-retryable = %#v, want %#v", activity.name, gotActivity.nonRetryable, activity.nonRetryable)
		}
	}
}

func TestGenerateExerciseRejectsMissingInput(t *testing.T) {
	execution := newWorkflowDouble(t)
	workflow := &exerciseWorkflow{workflow: execution.exec}

	_, customErr := workflow.GenerateExercise(context.Background(), nil)
	if customErr == nil || customErr.ErrType != apperror.ErrBadRequest {
		t.Fatalf("nil request: got %#v, want bad request", customErr)
	}
	if execution.started {
		t.Fatal("nil request started a workflow")
	}

	for _, tc := range []struct {
		name string
		ctx  context.Context
	}{
		{name: "missing", ctx: context.Background()},
		{name: "zero", ctx: context.WithValue(context.Background(), enums.UserID, uint64(0))},
		{name: "wrong type", ctx: context.WithValue(context.Background(), enums.UserID, "4")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, customErr := workflow.GenerateExercise(tc.ctx, &requests.GenerateExerciseRequest{IdempotencyKey: "key"})
			if customErr == nil || customErr.ErrType != apperror.ErrUnauthorized {
				t.Fatalf("got %#v, want unauthorized", customErr)
			}
			if execution.started {
				t.Fatal("missing user started a workflow")
			}
		})
	}
}

func TestGenerateExerciseStartFailure(t *testing.T) {
	execution := newWorkflowDouble(t)
	execution.allowStart("", errors.New("temporal down"))
	workflow := &exerciseWorkflow{workflow: execution.exec}

	_, customErr := workflow.GenerateExercise(userContext(4), &requests.GenerateExerciseRequest{IdempotencyKey: "key"})
	if customErr == nil || customErr.ErrType != apperror.ErrInternalServer {
		t.Fatalf("got %#v, want internal server", customErr)
	}
	if !errors.Is(customErr, ErrWorkflowNotStarted) {
		t.Fatalf("start failure was not marked as not started: %v", customErr)
	}
	if execution.startOpts.WorkflowID != "exercise_workflow-key" {
		t.Fatalf("workflow id = %q", execution.startOpts.WorkflowID)
	}
}

func TestGenerateExerciseQueryFailure(t *testing.T) {
	restoreQuery(t)
	queryErr := errors.New("query failed")
	wfQuery = func(context.Context, temporal_client.Temporal, string, string, string, temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
		return nil, queryErr
	}
	execution := newWorkflowDouble(t)
	execution.allowStart("", nil)
	workflow := &exerciseWorkflow{workflow: execution.exec}

	_, customErr := workflow.GenerateExercise(userContext(4), &requests.GenerateExerciseRequest{IdempotencyKey: "key"})
	if customErr == nil || customErr.ErrType != apperror.ErrInternalServer || !errors.Is(customErr, queryErr) {
		t.Fatalf("got %#v, want query failure", customErr)
	}
	if errors.Is(customErr, ErrWorkflowNotStarted) {
		t.Fatal("query failure was reported as a start failure")
	}
}

func TestGenerateExerciseRejectsUnreadyResult(t *testing.T) {
	for _, tc := range []struct {
		name   string
		result temporal_client.ExecutionData
	}{
		{name: "wrong type", result: temporalmocks.NewMockExecutionData(t)},
		{name: "missing exercise", result: &ExecutionData{}},
		{name: "empty status", result: &ExecutionData{GenerateExercise: &responses.GenerateExerciseResponse{}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			restoreQuery(t)
			wfQuery = func(context.Context, temporal_client.Temporal, string, string, string, temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
				return tc.result, nil
			}
			execution := newWorkflowDouble(t)
			execution.allowStart("", nil)
			workflow := &exerciseWorkflow{workflow: execution.exec}
			_, customErr := workflow.GenerateExercise(userContext(4), &requests.GenerateExerciseRequest{IdempotencyKey: "key"})
			if customErr == nil || customErr.ErrType != apperror.ErrInternalServer {
				t.Fatalf("got %#v, want internal server", customErr)
			}
		})
	}
}

func TestGenerateExerciseReturnsQueriedResponse(t *testing.T) {
	restoreQuery(t)
	want := &responses.GenerateExerciseResponse{ID: 8, Status: exerciseStatusProcessing, JLPTLevel: "3"}
	var queryType, workflowID, runID string
	wfQuery = func(_ context.Context, _ temporal_client.Temporal, gotQuery string, gotWorkflowID string, gotRunID string, _ temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
		queryType = gotQuery
		workflowID = gotWorkflowID
		runID = gotRunID
		return &ExecutionData{GenerateExercise: want}, nil
	}
	execution := newWorkflowDouble(t)
	execution.allowStart("run-9", nil)
	workflow := &exerciseWorkflow{workflow: execution.exec}

	got, customErr := workflow.GenerateExercise(userContext(4), &requests.GenerateExerciseRequest{IdempotencyKey: "key-9", JLPTNLevel: 3})
	if customErr != nil {
		t.Fatal(customErr)
	}
	if got != want {
		t.Fatalf("response = %#v", got)
	}
	if queryType != queryGenerateExerciseResponse || workflowID != "exercise_workflow-key-9" || runID != "run-9" {
		t.Fatalf("query %s workflow %s run %s", queryType, workflowID, runID)
	}
}

func TestGenerateExerciseWorkflowRejectsInput(t *testing.T) {
	workflow := &exerciseWorkflow{}
	if _, err := workflow.generateExercise(nil, nil, 1); err == nil || !strings.Contains(err.Error(), "request is nil") {
		t.Fatalf("nil request: %v", err)
	}
	if _, err := workflow.generateExercise(nil, &requests.GenerateExerciseRequest{IdempotencyKey: "key"}, 0); err == nil || !strings.Contains(err.Error(), "user id is missing") {
		t.Fatalf("zero user: %v", err)
	}
}

func TestGenerateExerciseWorkflowSuccess(t *testing.T) {
	restoreWorkflowFuncs(t)
	var seen *ExecutionData
	wfInit = func(_ workflow.Context, _ temporal_client.WorkflowExecution, queryType string, data temporal_client.ExecutionData, ready func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error) {
		if queryType != queryGenerateExerciseResponse {
			return nil, errors.New("unexpected query")
		}
		if err := ready(data); err != nil {
			return nil, err
		}
		seen = data.(*ExecutionData)
		return data, nil
	}
	wfExecute = func(workflow.Context, temporal_client.WorkflowExecution, temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
		return nil, nil
	}

	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow((&exerciseWorkflow{}).generateExercise, &requests.GenerateExerciseRequest{IdempotencyKey: "key", JLPTNLevel: 2}, uint64(4))
	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatal(err)
	}
	var response *responses.GenerateExerciseResponse
	if err := env.GetWorkflowResult(&response); err != nil {
		t.Fatal(err)
	}
	if response == nil || response.Status != exerciseStatusProcessing || response.JLPTLevel != "2" {
		t.Fatalf("response = %#v", response)
	}
	if seen == nil || seen.UserID != 4 || seen.IdempotencyKey != "key" || seen.CurrActivity != readPromptTxtActivity || seen.WorkflowID == "" || seen.RunID == "" {
		t.Fatalf("execution data = %#v", seen)
	}
}

func TestGenerateExerciseWorkflowInitAndExecuteErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		init    func(workflow.Context, temporal_client.WorkflowExecution, string, temporal_client.ExecutionData, func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error)
		execute func(workflow.Context, temporal_client.WorkflowExecution, temporal_client.ExecutionData) (temporal_client.ExecutionData, error)
		want    string
	}{
		{
			name: "not ready",
			init: func(_ workflow.Context, _ temporal_client.WorkflowExecution, _ string, _ temporal_client.ExecutionData, ready func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error) {
				if err := ready(&ExecutionData{}); err != nil {
					return nil, err
				}
				return &ExecutionData{}, nil
			},
			want: "not ready",
		},
		{
			name: "ready type",
			init: func(_ workflow.Context, _ temporal_client.WorkflowExecution, _ string, _ temporal_client.ExecutionData, ready func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error) {
				if err := ready(temporalmocks.NewMockExecutionData(t)); err != nil {
					return nil, err
				}
				return &ExecutionData{}, nil
			},
			want: "not ready",
		},
		{
			name: "init error",
			init: func(workflow.Context, temporal_client.WorkflowExecution, string, temporal_client.ExecutionData, func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error) {
				return nil, errors.New("init failed")
			},
			want: "init failed",
		},
		{
			name: "convert",
			init: func(workflow.Context, temporal_client.WorkflowExecution, string, temporal_client.ExecutionData, func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error) {
				return temporalmocks.NewMockExecutionData(t), nil
			},
			want: "convert result",
		},
		{
			name: "execute",
			init: func(_ workflow.Context, _ temporal_client.WorkflowExecution, _ string, data temporal_client.ExecutionData, _ func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error) {
				return data, nil
			},
			execute: func(workflow.Context, temporal_client.WorkflowExecution, temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
				return nil, errors.New("execute failed")
			},
			want: "execute failed",
		},
		{
			name: "missing response",
			init: func(_ workflow.Context, _ temporal_client.WorkflowExecution, _ string, data temporal_client.ExecutionData, _ func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error) {
				return data, nil
			},
			execute: func(_ workflow.Context, _ temporal_client.WorkflowExecution, data temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
				data.(*ExecutionData).GenerateExercise = nil
				return data, nil
			},
			want: "generated exercise is missing",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			restoreWorkflowFuncs(t)
			wfInit = tc.init
			if tc.execute != nil {
				wfExecute = tc.execute
			}
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.ExecuteWorkflow((&exerciseWorkflow{}).generateExercise, &requests.GenerateExerciseRequest{IdempotencyKey: "key", JLPTNLevel: 1}, uint64(4))
			if !env.IsWorkflowCompleted() {
				t.Fatal("workflow did not complete")
			}
			err := env.GetWorkflowError()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %q", err, tc.want)
			}
		})
	}
}

func restoreQuery(t *testing.T) {
	t.Helper()
	original := wfQuery
	t.Cleanup(func() { wfQuery = original })
}

func restoreWorkflowFuncs(t *testing.T) {
	t.Helper()
	originalInit := wfInit
	originalExecute := wfExecute
	t.Cleanup(func() {
		wfInit = originalInit
		wfExecute = originalExecute
	})
}

func userContext(userID uint64) context.Context {
	return context.WithValue(context.Background(), enums.UserID, userID)
}
