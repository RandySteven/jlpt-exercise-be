package onboarding_workflow

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
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func TestNewOnboardingWorkflowRegistersPipeline(t *testing.T) {
	execution := &fakeWorkflowExecution{}
	got := NewOnboardingWorkflow(execution, nil, nil, nil, nil, nil)
	if got == nil {
		t.Fatal("workflow was nil")
	}
	if len(execution.workflows) != 1 || execution.workflows[0] != onboardingWorkflowExecution {
		t.Fatalf("registered workflows = %#v", execution.workflows)
	}

	want := []recordedActivity{
		{name: persistOnboardingRequestActivity, signal: sgNoEvent, next: []string{registerUserActivity}, timeout: 25 * time.Second, attempts: 5, nonRetryable: []string{onboardingInputErrorType}},
		{name: registerUserActivity, signal: sgNoEvent, next: []string{sendActivationEmailActivity}, timeout: 25 * time.Second, attempts: 5, nonRetryable: []string{onboardingInputErrorType}},
		{name: sendActivationEmailActivity, signal: sgNoEvent, next: []string{updateOnboardingStatusActivity}, timeout: 25 * time.Second, attempts: 5, nonRetryable: []string{onboardingInputErrorType}},
		{name: updateOnboardingStatusActivity, signal: sgActivatedUserEvent, timeout: 25 * time.Second, attempts: 5, nonRetryable: []string{onboardingInputErrorType}},
	}
	if len(execution.activities) != len(want) {
		t.Fatalf("registered %d activities, want %d", len(execution.activities), len(want))
	}
	for i, activity := range want {
		gotActivity := execution.activities[i]
		if gotActivity.name != activity.name || gotActivity.signal != activity.signal || gotActivity.timeout != activity.timeout || gotActivity.attempts != activity.attempts {
			t.Fatalf("activity %d = %#v, want %#v", i, gotActivity, activity)
		}
		if strings.Join(gotActivity.next, ",") != strings.Join(activity.next, ",") || strings.Join(gotActivity.nonRetryable, ",") != strings.Join(activity.nonRetryable, ",") {
			t.Fatalf("activity %s = %#v, want %#v", activity.name, gotActivity, activity)
		}
	}
}

func TestOnboardingRegisterWorkflow(t *testing.T) {
	t.Run("start failure", func(t *testing.T) {
		startErr := errors.New("temporal down")
		execution := &fakeWorkflowExecution{startErr: startErr}
		workflow := &onboardingWorkflow{workflow: execution}
		_, customErr := workflow.OnboardingRegisterWorkflow(requestContext("req-1"), &requests.RegisterRequest{Email: "a@b.c"})
		if customErr == nil || customErr.ErrType != apperror.ErrInternalServer || !errors.Is(customErr, startErr) {
			t.Fatalf("got %#v", customErr)
		}
		if execution.startOpts.WorkflowID != "onboarding_workflow_execution-req-1" {
			t.Fatalf("workflow id = %q", execution.startOpts.WorkflowID)
		}
	})

	t.Run("query failure while workflow failed", func(t *testing.T) {
		restoreQuery(t)
		queryErr := errors.New("query failed")
		wfQueryGet = func(context.Context, temporal_client.Temporal, string, string, string, temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
			return nil, queryErr
		}
		execution := &fakeWorkflowExecution{status: temporal_client.StatusFailed}
		workflow := &onboardingWorkflow{workflow: execution}
		_, customErr := workflow.OnboardingRegisterWorkflow(requestContext("req-2"), &requests.RegisterRequest{})
		if customErr == nil || !strings.Contains(customErr.Error(), temporal_client.StatusFailed) {
			t.Fatalf("got %#v", customErr)
		}
	})

	t.Run("query failure without workflow failure", func(t *testing.T) {
		restoreQuery(t)
		queryErr := errors.New("query failed")
		wfQueryGet = func(context.Context, temporal_client.Temporal, string, string, string, temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
			return nil, queryErr
		}
		workflow := &onboardingWorkflow{workflow: &fakeWorkflowExecution{statusErr: errors.New("status unavailable")}}
		_, customErr := workflow.OnboardingRegisterWorkflow(requestContext("req-3"), &requests.RegisterRequest{})
		if customErr == nil || !errors.Is(customErr, queryErr) {
			t.Fatalf("got %#v", customErr)
		}
	})

	t.Run("unusable result", func(t *testing.T) {
		for _, result := range []temporal_client.ExecutionData{&otherExecutionData{}, &ExecutionData{}} {
			restoreQuery(t)
			wfQueryGet = func(context.Context, temporal_client.Temporal, string, string, string, temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
				return result, nil
			}
			workflow := &onboardingWorkflow{workflow: &fakeWorkflowExecution{}}
			_, customErr := workflow.OnboardingRegisterWorkflow(requestContext("req-4"), &requests.RegisterRequest{})
			if customErr == nil || customErr.ErrType != apperror.ErrInternalServer || !strings.Contains(customErr.Error(), "register response is missing") {
				t.Fatalf("got %#v", customErr)
			}
		}
	})

	t.Run("success", func(t *testing.T) {
		restoreQuery(t)
		want := &responses.RegisterResponse{UserID: 4, ActivationToken: "token"}
		var queryType, workflowID, runID string
		wfQueryGet = func(_ context.Context, _ temporal_client.Temporal, gotQuery string, gotWorkflowID string, gotRunID string, _ temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
			queryType = gotQuery
			workflowID = gotWorkflowID
			runID = gotRunID
			return &ExecutionData{Response: want}, nil
		}
		workflow := &onboardingWorkflow{workflow: &fakeWorkflowExecution{runID: "run-8"}}
		got, customErr := workflow.OnboardingRegisterWorkflow(requestContext("req-8"), &requests.RegisterRequest{Email: "a@b.c"})
		if customErr != nil {
			t.Fatal(customErr)
		}
		if got != want || queryType != queryRegisterResponse || workflowID != "onboarding_workflow_execution-req-8" || runID != "run-8" {
			t.Fatalf("response %#v query %s workflow %s run %s", got, queryType, workflowID, runID)
		}
	})
}

func TestRegisterWorkflowFailure(t *testing.T) {
	workflow := &onboardingWorkflow{workflow: &fakeWorkflowExecution{statusErr: errors.New("unavailable")}}
	if err := workflow.registerWorkflowFailure("wf", "run"); err != nil {
		t.Fatalf("status lookup failure returned %v", err)
	}

	workflow.workflow = &fakeWorkflowExecution{status: temporal_client.StatusFailed}
	if err := workflow.registerWorkflowFailure("wf", "run"); err == nil || !strings.Contains(err.Error(), temporal_client.StatusFailed) {
		t.Fatalf("failed status: %v", err)
	}

	workflow.workflow = &fakeWorkflowExecution{status: temporal_client.StatusRejected}
	if err := workflow.registerWorkflowFailure("wf", "run"); err == nil || !strings.Contains(err.Error(), temporal_client.StatusRejected) {
		t.Fatalf("rejected status: %v", err)
	}

	workflow.workflow = &fakeWorkflowExecution{status: "RUNNING"}
	if err := workflow.registerWorkflowFailure("wf", "run"); err != nil {
		t.Fatal(err)
	}
}

func TestCancelIfRegisterFailed(t *testing.T) {
	t.Run("context already done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		(&onboardingWorkflow{workflow: &fakeWorkflowExecution{}}).cancelIfRegisterFailed(ctx, func() {
			t.Fatal("cancel was called after the context finished")
		}, "wf", "run")
	})

	t.Run("status lookup keeps waiting", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		called := false
		(&onboardingWorkflow{workflow: &fakeWorkflowExecution{statusErr: errors.New("unavailable")}}).cancelIfRegisterFailed(ctx, func() {
			called = true
		}, "wf", "run")
		if called {
			t.Fatal("status lookup failure cancelled the wait")
		}
	})

	for _, status := range []string{temporal_client.StatusFailed, temporal_client.StatusRejected} {
		t.Run(status, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			go func() {
				(&onboardingWorkflow{workflow: &fakeWorkflowExecution{status: status}}).cancelIfRegisterFailed(ctx, cancel, "wf", "run")
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("failed status did not cancel the wait")
			}
			if ctx.Err() == nil {
				t.Fatal("context stayed open")
			}
		})
	}
}

func TestActivateUserRegisterWorkflow(t *testing.T) {
	workflow := &onboardingWorkflow{workflow: &fakeWorkflowExecution{}}
	_, customErr := workflow.ActivateUserRegisterWorkflow(context.Background(), &requests.ActivateRequest{Token: "not-a-token"})
	if customErr == nil || customErr.ErrType != apperror.ErrInternalServer {
		t.Fatalf("invalid token: %#v", customErr)
	}

	token := activationTokenGenerate(4, 9, "wf-activate", "run-activate")
	signalErr := errors.New("signal failed")
	execution := &fakeWorkflowExecution{signalErr: signalErr}
	workflow.workflow = execution
	_, customErr = workflow.ActivateUserRegisterWorkflow(context.Background(), &requests.ActivateRequest{Token: token})
	if customErr == nil || !errors.Is(customErr, signalErr) {
		t.Fatalf("signal failure: %#v", customErr)
	}
	if execution.signaledID != "wf-activate" || execution.signaledRun != "run-activate" || execution.signalName != sgActivatedUserEvent || execution.signalArg != true {
		t.Fatalf("signal = %#v", execution)
	}

	resultErr := errors.New("result failed")
	workflow.workflow = &fakeWorkflowExecution{resultErr: resultErr}
	_, customErr = workflow.ActivateUserRegisterWorkflow(context.Background(), &requests.ActivateRequest{Token: token})
	if customErr == nil || !errors.Is(customErr, resultErr) {
		t.Fatalf("result failure: %#v", customErr)
	}

	workflow.workflow = &fakeWorkflowExecution{}
	got, customErr := workflow.ActivateUserRegisterWorkflow(context.Background(), &requests.ActivateRequest{Token: token})
	if customErr != nil {
		t.Fatal(customErr)
	}
	if got.Status != "success" || !got.Activate {
		t.Fatalf("response = %#v", got)
	}
}

func TestRegisterWorkflowFunction(t *testing.T) {
	if _, err := (&onboardingWorkflow{}).register(nil, "wf", nil); err == nil || !strings.Contains(err.Error(), "request is nil") {
		t.Fatal(err)
	}

	t.Run("success", func(t *testing.T) {
		restoreWorkflowFuncs(t)
		cache := &fakeOnboardingCache{}
		var seen *ExecutionData
		wfInit = func(_ workflow.Context, _ temporal_client.WorkflowExecution, queryType string, data temporal_client.ExecutionData, ready func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error) {
			if queryType != queryRegisterResponse {
				return nil, errors.New("unexpected query")
			}
			current := data.(*ExecutionData)
			current.Response = &responses.RegisterResponse{ActivationToken: "token", OnboardingID: 3}
			current.Request = &requests.RegisterRequest{Email: "a@b.c"}
			if err := ready(current); err != nil {
				return nil, err
			}
			seen = current
			return current, nil
		}
		wfExecute = func(workflow.Context, temporal_client.WorkflowExecution, temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
			return nil, nil
		}

		var suite testsuite.WorkflowTestSuite
		env := suite.NewTestWorkflowEnvironment()
		workflow := &onboardingWorkflow{onboardingCache: cache}
		env.ExecuteWorkflow(workflow.register, "wf-register", &requests.RegisterRequest{Email: "a@b.c"})
		if err := env.GetWorkflowError(); err != nil {
			t.Fatal(err)
		}
		var response *responses.RegisterResponse
		if err := env.GetWorkflowResult(&response); err != nil {
			t.Fatal(err)
		}
		if response == nil || response.ActivationToken != "token" || seen == nil || seen.WorkflowID != "wf-register" || seen.RunID == "" {
			t.Fatalf("response %#v execution %#v", response, seen)
		}
		if cache.email != "a@b.c" || cache.response == nil || cache.response.ActivationToken != response.ActivationToken {
			t.Fatalf("cached %#v for %s", cache.response, cache.email)
		}
	})

	t.Run("ready callback", func(t *testing.T) {
		restoreWorkflowFuncs(t)
		cache := &fakeOnboardingCache{setErr: errors.New("redis down")}
		wfInit = func(_ workflow.Context, _ temporal_client.WorkflowExecution, _ string, data temporal_client.ExecutionData, ready func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error) {
			if err := ready(&otherExecutionData{}); err == nil || !strings.Contains(err.Error(), "not ready") {
				return nil, errors.New("bad type was ready")
			}
			if err := ready(&ExecutionData{Request: &requests.RegisterRequest{Email: "a@b.c"}}); err == nil {
				return nil, errors.New("empty response was ready")
			}
			current := data.(*ExecutionData)
			current.Request = &requests.RegisterRequest{Email: "a@b.c"}
			current.Response = &responses.RegisterResponse{OnboardingID: 1}
			if err := ready(current); err != nil {
				return nil, err
			}
			return current, nil
		}
		wfExecute = func(workflow.Context, temporal_client.WorkflowExecution, temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
			return nil, nil
		}
		var suite testsuite.WorkflowTestSuite
		env := suite.NewTestWorkflowEnvironment()
		env.ExecuteWorkflow((&onboardingWorkflow{onboardingCache: cache}).register, "wf", &requests.RegisterRequest{Email: "a@b.c"})
		if err := env.GetWorkflowError(); err != nil {
			t.Fatal(err)
		}
		if cache.email != "a@b.c" {
			t.Fatalf("cache email = %q", cache.email)
		}
	})

	for _, tc := range []struct {
		name    string
		init    func(workflow.Context, temporal_client.WorkflowExecution, string, temporal_client.ExecutionData, func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error)
		execute func(workflow.Context, temporal_client.WorkflowExecution, temporal_client.ExecutionData) (temporal_client.ExecutionData, error)
		want    string
	}{
		{
			name: "init",
			init: func(workflow.Context, temporal_client.WorkflowExecution, string, temporal_client.ExecutionData, func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error) {
				return nil, errors.New("init failed")
			},
			want: "init failed",
		},
		{
			name: "convert",
			init: func(workflow.Context, temporal_client.WorkflowExecution, string, temporal_client.ExecutionData, func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error) {
				return &otherExecutionData{}, nil
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			restoreWorkflowFuncs(t)
			wfInit = tc.init
			if tc.execute != nil {
				wfExecute = tc.execute
			}
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.ExecuteWorkflow((&onboardingWorkflow{}).register, "wf", &requests.RegisterRequest{Email: "a@b.c"})
			err := env.GetWorkflowError()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %q", err, tc.want)
			}
		})
	}
}

func restoreQuery(t *testing.T) {
	t.Helper()
	original := wfQueryGet
	t.Cleanup(func() { wfQueryGet = original })
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

func requestContext(requestID string) context.Context {
	return context.WithValue(context.Background(), enums.RequestID, requestID)
}
