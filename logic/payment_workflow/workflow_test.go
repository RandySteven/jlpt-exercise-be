package payment_workflow

import (
	"context"
	"errors"
	"testing"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/testsuite"
	temporalworkflow "go.temporal.io/sdk/workflow"
)

func TestNewPaymentWorkflowRegistersPipeline(t *testing.T) {
	execution := &fakeWorkflowExecution{}
	got := NewPaymentWorkflow(execution, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if got == nil {
		t.Fatal("workflow was nil")
	}
	if len(execution.workflows) != 1 || execution.workflows[0] != paymentWorkflowName {
		t.Fatalf("registered workflows = %#v", execution.workflows)
	}

	want := []recordedActivity{
		{name: checkPlanActivityName, signal: sgCreateSubscription, next: []string{persistPaymentActivityName, planExistActivityName}, timeout: 25 * time.Second, attempts: 5, nonRetryable: nonRetryableErrorTypes},
		{name: planExistActivityName, signal: sgNoNeed, timeout: 25 * time.Second, attempts: 5, nonRetryable: nonRetryableErrorTypes},
		{name: persistPaymentActivityName, signal: sgNoNeed, next: []string{createSubscriptionActivityName}, timeout: 25 * time.Second, attempts: 5, nonRetryable: nonRetryableErrorTypes},
		{name: createSubscriptionActivityName, signal: sgNoNeed, next: []string{midtransCallActivityName}, timeout: 25 * time.Second, attempts: 5, nonRetryable: nonRetryableErrorTypes},
		{name: midtransCallActivityName, signal: sgNoNeed, next: []string{updatePaymentStatusActivityName, paymentFailedActivityName}, timeout: 25 * time.Second, attempts: 5, nonRetryable: nonRetryableErrorTypes},
		{name: paymentFailedActivityName, signal: sgNoNeed, timeout: 25 * time.Second, attempts: 5, nonRetryable: nonRetryableErrorTypes},
		{name: updatePaymentStatusActivityName, signal: sgConfirmPayment, next: []string{publishPaymentActivityName}, timeout: 25 * time.Second, attempts: 5, nonRetryable: nonRetryableErrorTypes},
		{name: publishPaymentActivityName, signal: sgNoNeed, next: []string{sendSubscriptionEmailActivityName}, timeout: 25 * time.Second, attempts: 5, nonRetryable: nonRetryableErrorTypes},
		{name: sendSubscriptionEmailActivityName, signal: sgNoNeed, next: []string{sgNoNeed}, timeout: 25 * time.Second, attempts: 5, nonRetryable: nonRetryableErrorTypes},
	}
	if len(execution.activities) != len(want) {
		t.Fatalf("registered %d activities, want %d", len(execution.activities), len(want))
	}
	for i, activity := range want {
		gotActivity := execution.activities[i]
		if gotActivity.name != activity.name || gotActivity.signal != activity.signal || gotActivity.timeout != activity.timeout || gotActivity.attempts != activity.attempts {
			t.Fatalf("activity %d = %#v, want %#v", i, gotActivity, activity)
		}
		if len(gotActivity.next) != len(activity.next) {
			t.Fatalf("activity %s next = %#v, want %#v", activity.name, gotActivity.next, activity.next)
		}
		for j := range activity.next {
			if gotActivity.next[j] != activity.next[j] {
				t.Fatalf("activity %s next = %#v, want %#v", activity.name, gotActivity.next, activity.next)
			}
		}
		if len(gotActivity.nonRetryable) != len(activity.nonRetryable) {
			t.Fatalf("activity %s non-retryable = %#v", activity.name, gotActivity.nonRetryable)
		}
	}
}

func TestCreateSubscriptionClient(t *testing.T) {
	execution := &fakeWorkflowExecution{}
	workflow := &paymentWorkflow{workflow: execution}

	_, err := workflow.CreateSubscription(context.Background(), nil)
	requireCustom(t, err, apperror.ErrBadRequest)
	_, err = workflow.CreateSubscription(userContext(4), &requests.CreateSubscriptionRequest{})
	requireCustom(t, err, apperror.ErrBadRequest)
	_, err = workflow.CreateSubscription(context.Background(), &requests.CreateSubscriptionRequest{IdempotencyKey: "key", PlanID: 3})
	requireCustom(t, err, apperror.ErrUnauthorized)
	if execution.started {
		t.Fatal("invalid request started a workflow")
	}

	execution.startErr = errors.New("temporal down")
	_, err = workflow.CreateSubscription(userContext(4), &requests.CreateSubscriptionRequest{IdempotencyKey: "key", PlanID: 3})
	requireCustom(t, err, apperror.ErrInternalServer)
	if execution.startOpts.WorkflowID != "payment_workflow-key" {
		t.Fatalf("workflow id = %q", execution.startOpts.WorkflowID)
	}

	restoreQuery(t)
	queryErr := errors.New("query failed")
	qryExec = func(context.Context, temporal_client.Temporal, string, string, string, temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
		return nil, queryErr
	}
	execution = &fakeWorkflowExecution{status: temporal_client.StatusFailed}
	workflow.workflow = execution
	_, err = workflow.CreateSubscription(userContext(4), &requests.CreateSubscriptionRequest{IdempotencyKey: "key", PlanID: 3})
	custom := requireCustom(t, err, apperror.ErrInternalServer)
	if custom.Error() == queryErr.Error() {
		t.Fatal("workflow failure was reported as the query error")
	}

	want := &responses.CreateSubscriptionResponse{SubscriptionID: "51"}
	var queryType, workflowID, runID string
	qryExec = func(_ context.Context, _ temporal_client.Temporal, gotQuery string, gotWorkflowID string, gotRunID string, _ temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
		queryType = gotQuery
		workflowID = gotWorkflowID
		runID = gotRunID
		return &ExecutionData{Response: want}, nil
	}
	execution = &fakeWorkflowExecution{runID: "run-9"}
	workflow.workflow = execution
	got, err := workflow.CreateSubscription(userContext(4), &requests.CreateSubscriptionRequest{IdempotencyKey: "key-9", PlanID: 3})
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("response = %#v", got)
	}
	if queryType != queryPaymentWorkflow || workflowID != "payment_workflow-key-9" || runID != "run-9" {
		t.Fatalf("query %s workflow %s run %s", queryType, workflowID, runID)
	}
	if execution.signalName != sgCreateSubscription || execution.signalArg != true {
		t.Fatalf("signal = %s %#v", execution.signalName, execution.signalArg)
	}

	execution = &fakeWorkflowExecution{startErr: &serviceerror.WorkflowExecutionAlreadyStarted{}}
	workflow.workflow = execution
	if _, err = workflow.CreateSubscription(userContext(4), &requests.CreateSubscriptionRequest{IdempotencyKey: "key-9", PlanID: 3}); err != nil {
		t.Fatal(err)
	}
	if execution.signalName != sgCreateSubscription || execution.signaledID != "payment_workflow-key-9" {
		t.Fatalf("resume signal = %s %s", execution.signalName, execution.signaledID)
	}
}

func TestPaymentConfirmation(t *testing.T) {
	workflow := &paymentWorkflow{workflow: &fakeWorkflowExecution{}}
	_, err := workflow.PaymentConfirmation(context.Background(), nil)
	requireCustom(t, err, apperror.ErrBadRequest)

	restoreQuery(t)
	qryExec = func(context.Context, temporal_client.Temporal, string, string, string, temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
		return &ExecutionData{Request: &requests.CreateSubscriptionRequest{PlanID: 9}}, nil
	}
	_, err = workflow.PaymentConfirmation(context.Background(), &requests.PaymentConfirmationRequest{IdempotencyKey: "key", PlanID: 3})
	requireCustom(t, err, apperror.ErrBadRequest)

	qryExec = func(context.Context, temporal_client.Temporal, string, string, string, temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
		return &ExecutionData{
			RunID:               "run-2",
			Request:             &requests.CreateSubscriptionRequest{IdempotencyKey: "key", PlanID: 3},
			MidtransTransaction: &midtrans_client.TransactionResponse{RedirectURL: "https://pay.example/snap"},
			Subscription:        &models.Subscription{Status: subscriptionStatusPending},
		}, nil
	}
	execution := &fakeWorkflowExecution{}
	workflow.workflow = execution
	got, err := workflow.PaymentConfirmation(context.Background(), &requests.PaymentConfirmationRequest{IdempotencyKey: "key", PlanID: 3})
	if err != nil {
		t.Fatal(err)
	}
	if got.RedirectURL != "https://pay.example/snap" || execution.signalName != sgConfirmPayment || execution.signaledRun != "run-2" {
		t.Fatalf("confirmation = %#v signal %s %s", got, execution.signalName, execution.signaledRun)
	}

	qryExec = func(context.Context, temporal_client.Temporal, string, string, string, temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
		return &ExecutionData{
			Request:             &requests.CreateSubscriptionRequest{PlanID: 3},
			MidtransTransaction: &midtrans_client.TransactionResponse{RedirectURL: "https://pay.example/snap"},
			Subscription:        &models.Subscription{Status: subscriptionStatusActive},
		}, nil
	}
	execution = &fakeWorkflowExecution{}
	workflow.workflow = execution
	got, err = workflow.PaymentConfirmation(context.Background(), &requests.PaymentConfirmationRequest{IdempotencyKey: "key", PlanID: 3})
	if err != nil || got.RedirectURL != "https://pay.example/snap" || execution.signalName != "" {
		t.Fatalf("active confirmation = %#v signal %q err %v", got, execution.signalName, err)
	}
}

func TestPaymentWorkflowFunction(t *testing.T) {
	workflow := &paymentWorkflow{}
	if _, err := workflow.payment(nil, nil, 4); err == nil || err.Error() != "subscription request is nil" {
		t.Fatalf("nil request: %v", err)
	}
	if _, err := workflow.payment(nil, &requests.CreateSubscriptionRequest{IdempotencyKey: "key"}, 0); err == nil || err.Error() != "user id is missing" {
		t.Fatalf("zero user: %v", err)
	}

	restoreWorkflowFuncs(t)
	var seen *ExecutionData
	wfInit = func(_ temporalworkflow.Context, _ temporal_client.WorkflowExecution, queryType string, data temporal_client.ExecutionData, _ func(temporal_client.ExecutionData) error) (temporal_client.ExecutionData, error) {
		if queryType != queryPaymentWorkflow {
			return nil, errors.New("unexpected query")
		}
		seen = data.(*ExecutionData)
		return data, nil
	}
	wfExecute = func(_ temporalworkflow.Context, _ temporal_client.WorkflowExecution, data temporal_client.ExecutionData) (temporal_client.ExecutionData, error) {
		data.(*ExecutionData).setSubscriptionResponse(51)
		return data, nil
	}

	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow((&paymentWorkflow{}).payment, &requests.CreateSubscriptionRequest{IdempotencyKey: "key", PlanID: 3}, uint64(4))
	if !env.IsWorkflowCompleted() || env.GetWorkflowError() != nil {
		t.Fatalf("workflow error: %v", env.GetWorkflowError())
	}
	var response *responses.CreateSubscriptionResponse
	if err := env.GetWorkflowResult(&response); err != nil {
		t.Fatal(err)
	}
	if response.SubscriptionID != "51" || seen.UserID != 4 || seen.CurrActivity != checkPlanActivityName || seen.WorkflowID == "" {
		t.Fatalf("response %#v data %#v", response, seen)
	}
}

func TestExecutionDataSignalPayloads(t *testing.T) {
	original := &ExecutionData{UserID: 4, Request: &requests.CreateSubscriptionRequest{IdempotencyKey: "keep"}}
	if err := original.Unmarshal([]byte("true")); err != nil || original.UserID != 4 || original.Request.IdempotencyKey != "keep" {
		t.Fatalf("bool signal changed state: %#v err %v", original, err)
	}
	if err := original.Unmarshal([]byte(`{"idempotency_key":"next","plan_id":3}`)); err != nil || original.Request.PlanID != 3 || original.UserID != 4 {
		t.Fatalf("request patch = %#v err %v", original, err)
	}

	clone := original.Clone().(*ExecutionData)
	clone.Request.PlanID = 9
	if original.Request.PlanID == 9 {
		t.Fatal("clone shared the request")
	}
}

func restoreQuery(t *testing.T) {
	t.Helper()
	original := qryExec
	t.Cleanup(func() { qryExec = original })
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

func userContext(id uint64) context.Context {
	return context.WithValue(context.Background(), enums.UserID, id)
}
