package onboarding_workflow

import (
	"context"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

type recordedActivity struct {
	name         string
	signal       string
	next         []string
	timeout      time.Duration
	attempts     int32
	nonRetryable []string
}

type fakeWorkflowExecution struct {
	activities []recordedActivity
	workflows  []string

	started   bool
	startOpts temporal_client.StartWorkflowOptions
	startErr  error
	runID     string

	signalErr   error
	signalName  string
	signalArg   any
	signaledID  string
	signaledRun string

	status     string
	statusErr  error
	resultErr  error
	workflowID string
}

func (f *fakeWorkflowExecution) Execute(workflow.Context, temporal_client.ExecutionData) error {
	return nil
}

func (f *fakeWorkflowExecution) AddTransitionActivityWithOptions(activityName string, signalEvent string, _ temporal_client.ActivityFunction, options *workflow.ActivityOptions, nextActivities ...string) {
	recorded := recordedActivity{name: activityName, signal: signalEvent, next: append([]string(nil), nextActivities...)}
	if options != nil {
		recorded.timeout = options.StartToCloseTimeout
		if options.RetryPolicy != nil {
			recorded.attempts = options.RetryPolicy.MaximumAttempts
			recorded.nonRetryable = append([]string(nil), options.RetryPolicy.NonRetryableErrorTypes...)
		}
	}
	f.activities = append(f.activities, recorded)
}

func (f *fakeWorkflowExecution) RegisterWorkflow(name string, _ interface{}) {
	f.workflows = append(f.workflows, name)
}

func (f *fakeWorkflowExecution) GetWorkflowExecutionData(workflow.Context, string, interface{}) error {
	return nil
}

func (f *fakeWorkflowExecution) StartWorkflow(_ context.Context, opts temporal_client.StartWorkflowOptions, _ interface{}, _ ...interface{}) (client.WorkflowRun, error) {
	f.started = true
	f.startOpts = opts
	if f.startErr != nil {
		return nil, f.startErr
	}
	id := f.runID
	if id == "" {
		id = "run-1"
	}
	return &fakeWorkflowRun{id: opts.WorkflowID, runID: id}, nil
}

func (f *fakeWorkflowExecution) GetWorkflowResult(context.Context, string, string, interface{}) error {
	return f.resultErr
}

func (f *fakeWorkflowExecution) StartChildWorkflow(workflow.Context, string, string, interface{}, interface{}) error {
	return nil
}

func (f *fakeWorkflowExecution) SignalWorkflow(_ context.Context, workflowID string, runID string, signalEvent string, arg interface{}) error {
	f.signaledID = workflowID
	f.signaledRun = runID
	f.signalName = signalEvent
	f.signalArg = arg
	return f.signalErr
}

func (f *fakeWorkflowExecution) Goroutine(workflow.Context, func(workflow.Context)) {}

func (f *fakeWorkflowExecution) GetSignalResult(workflow.Context, string, interface{}) error {
	return nil
}

func (f *fakeWorkflowExecution) SignalExternalWorkflow(workflow.Context, string, string, string, interface{}) error {
	return nil
}

func (f *fakeWorkflowExecution) GetExternalWorkflowResult(workflow.Context, string, string, interface{}) error {
	return nil
}

func (f *fakeWorkflowExecution) WaitForSignal(workflow.Context, string, interface{}, time.Duration) error {
	return nil
}

func (f *fakeWorkflowExecution) ListenSignal(workflow.Context, string, interface{}, func(workflow.Context)) error {
	return nil
}

func (f *fakeWorkflowExecution) WaitForAnySignal(workflow.Context, map[string]interface{}) (string, error) {
	return "", nil
}

func (f *fakeWorkflowExecution) SetQueryHandler(workflow.Context, string, interface{}) error {
	return nil
}

func (f *fakeWorkflowExecution) SetUpdateHandlerWithOptions(workflow.Context, string, interface{}, workflow.UpdateHandlerOptions) error {
	return nil
}

func (f *fakeWorkflowExecution) UpdateWorkflow(context.Context, string, string, string, client.WorkflowUpdateStage, interface{}) (client.WorkflowUpdateHandle, error) {
	return nil, nil
}

func (f *fakeWorkflowExecution) GetWorkflowStatus(context.Context, string, string) (string, error) {
	if f.statusErr != nil {
		return "", f.statusErr
	}
	if f.status == "" {
		return "RUNNING", nil
	}
	return f.status, nil
}

func (f *fakeWorkflowExecution) GetWorkflowID() string { return f.workflowID }

func (f *fakeWorkflowExecution) SetWorkflowID(workflowID string) { f.workflowID = workflowID }

func (f *fakeWorkflowExecution) SetExecutionData(temporal_client.ExecutionData) {}

func (f *fakeWorkflowExecution) GetExecutionData() temporal_client.ExecutionData { return nil }

type fakeWorkflowRun struct {
	id    string
	runID string
}

func (f *fakeWorkflowRun) GetID() string    { return f.id }
func (f *fakeWorkflowRun) GetRunID() string { return f.runID }
func (f *fakeWorkflowRun) Get(context.Context, interface{}) error {
	return nil
}
func (f *fakeWorkflowRun) GetWithOptions(context.Context, interface{}, client.WorkflowRunGetOptions) error {
	return nil
}

type otherExecutionData struct{}

func (o *otherExecutionData) Clone() temporal_client.ExecutionData { return o }
func (o *otherExecutionData) GetActivity() string                  { return "" }
func (o *otherExecutionData) Marshal() ([]byte, error)             { return []byte("{}"), nil }
func (o *otherExecutionData) SetActivity(string)                   {}
func (o *otherExecutionData) Unmarshal([]byte) error               { return nil }

type fakeOnboardingRepository struct {
	saved     *models.Onboarding
	saveErr   error
	updated   *models.Onboarding
	updateErr error
	nextID    uint64
}

func (f *fakeOnboardingRepository) Save(_ context.Context, entity *models.Onboarding) (*models.Onboarding, error) {
	if f.saveErr != nil {
		return nil, f.saveErr
	}
	copied := *entity
	if copied.ID == 0 {
		if f.nextID == 0 {
			f.nextID = 31
		}
		copied.ID = f.nextID
	}
	f.saved = &copied
	return f.saved, nil
}

func (f *fakeOnboardingRepository) Update(_ context.Context, entity *models.Onboarding) (*models.Onboarding, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	copied := *entity
	f.updated = &copied
	return f.updated, nil
}

func (f *fakeOnboardingRepository) FindByEmail(context.Context, string) (*models.Onboarding, error) {
	return nil, nil
}

type fakeUserRepository struct {
	saved     *models.User
	saveErr   error
	updated   *models.User
	updateErr error
}

func (f *fakeUserRepository) Save(_ context.Context, entity *models.User) (*models.User, error) {
	if f.saveErr != nil {
		return nil, f.saveErr
	}
	copied := *entity
	if copied.ID == 0 {
		copied.ID = 11
	}
	f.saved = &copied
	return f.saved, nil
}

func (f *fakeUserRepository) FindByID(context.Context, uint64) (*models.User, error) {
	return nil, nil
}

func (f *fakeUserRepository) FindAll(context.Context, uint64, uint64) ([]*models.User, error) {
	return nil, nil
}

func (f *fakeUserRepository) Update(_ context.Context, entity *models.User) (*models.User, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	copied := *entity
	f.updated = &copied
	return f.updated, nil
}

func (f *fakeUserRepository) DeleteByID(context.Context, uint64) error { return nil }

type fakeOnboardingCache struct {
	setErr   error
	email    string
	response *responses.RegisterResponse
}

func (f *fakeOnboardingCache) Get(context.Context, string) (interface{}, error) { return nil, nil }

func (f *fakeOnboardingCache) Set(context.Context, string, interface{}) error { return nil }

func (f *fakeOnboardingCache) GetLoginAttempt(context.Context, string) (int, error) { return 0, nil }

func (f *fakeOnboardingCache) SetLoginAttempt(context.Context, string, int) error { return nil }

func (f *fakeOnboardingCache) SetRegisterResponse(_ context.Context, email string, response *responses.RegisterResponse) error {
	f.email = email
	f.response = response
	return f.setErr
}

func (f *fakeOnboardingCache) GetRegisterResponse(context.Context, string) (*responses.RegisterResponse, error) {
	return nil, nil
}
