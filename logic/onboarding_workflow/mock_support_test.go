package onboarding_workflow

import (
	"context"
	"testing"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	temporalmocks "github.com/RandySteven/jlpt-exercise-be/mocks/temporal"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

type recordedActivity struct {
	name         string
	signal       string
	next         []string
	timeout      time.Duration
	attempts     int
	nonRetryable []string
}
type workflowDouble struct {
	exec        *temporalmocks.MockWorkflowExecution
	activities  []recordedActivity
	workflows   []string
	started     bool
	startOpts   temporal_client.StartWorkflowOptions
	signalName  string
	signalArg   any
	signaledID  string
	signaledRun string
}

func newWorkflowDouble(t *testing.T) *workflowDouble {
	t.Helper()
	double := &workflowDouble{exec: temporalmocks.NewMockWorkflowExecution(t)}
	record := func(name string, signal string, _ temporal_client.ActivityFunction, options *workflow.ActivityOptions, next ...string) {
		recorded := recordedActivity{name: name, signal: signal, next: append([]string(nil), next...)}
		if options != nil {
			recorded.timeout = options.StartToCloseTimeout
			if options.RetryPolicy != nil {
				recorded.attempts = int(options.RetryPolicy.MaximumAttempts)
				recorded.nonRetryable = append([]string(nil), options.RetryPolicy.NonRetryableErrorTypes...)
			}
		}
		double.activities = append(double.activities, recorded)
	}
	double.exec.EXPECT().
		AddTransitionActivityWithOptions(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(record).Return().Maybe()
	double.exec.EXPECT().
		AddTransitionActivityWithOptions(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(record).Return().Maybe()
	double.exec.EXPECT().
		RegisterWorkflow(mock.Anything, mock.Anything).
		Run(func(name string, _ interface{}) {
			double.workflows = append(double.workflows, name)
		}).
		Return().Maybe()
	return double
}

func (w *workflowDouble) allowStart(runID string, startErr error) {
	w.exec.EXPECT().
		StartWorkflow(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, opts temporal_client.StartWorkflowOptions, _ interface{}, _ ...interface{}) (client.WorkflowRun, error) {
			w.started = true
			w.startOpts = opts
			if startErr != nil {
				return nil, startErr
			}
			id := runID
			if id == "" {
				id = "run-1"
			}
			return workflowRun{id: opts.WorkflowID, runID: id}, nil
		})
}

func (w *workflowDouble) allowStatus(status string, statusErr error) {
	w.exec.EXPECT().
		GetWorkflowStatus(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(context.Context, string, string) (string, error) {
			if statusErr != nil {
				return "", statusErr
			}
			if status == "" {
				return "RUNNING", nil
			}
			return status, nil
		}).Maybe()
}

func (w *workflowDouble) allowSignal(signalErr error) {
	w.exec.EXPECT().
		SignalWorkflow(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ context.Context, workflowID string, runID string, signalEvent string, arg interface{}) {
			w.signaledID = workflowID
			w.signaledRun = runID
			w.signalName = signalEvent
			w.signalArg = arg
		}).
		Return(signalErr).Maybe()
}

// workflowRun implements client.WorkflowRun. Mockery cannot generate this mock:
// the SDK aliases the interface to go.temporal.io/sdk/internal, which external
// packages are not allowed to import.
type workflowRun struct {
	id    string
	runID string
}

func (r workflowRun) GetID() string    { return r.id }
func (r workflowRun) GetRunID() string { return r.runID }
func (r workflowRun) Get(context.Context, interface{}) error {
	return nil
}
func (r workflowRun) GetWithOptions(context.Context, interface{}, client.WorkflowRunGetOptions) error {
	return nil
}

func (w *workflowDouble) allowResult(resultErr error) {
	w.exec.EXPECT().
		GetWorkflowResult(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(resultErr).Maybe()
}
