package exercise_workflow

import (
	"context"
	"database/sql"
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

	signalErr  error
	signalName string
	signalArg  any
	signaledID string

	status    string
	statusErr error
	resultErr error

	executionData temporal_client.ExecutionData
	workflowID    string
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

func (f *fakeWorkflowExecution) StartWorkflow(_ context.Context, opts temporal_client.StartWorkflowOptions, workflowFn interface{}, _ ...interface{}) (client.WorkflowRun, error) {
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

func (f *fakeWorkflowExecution) SignalWorkflow(_ context.Context, workflowID string, _ string, signalEvent string, arg interface{}) error {
	f.signaledID = workflowID
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

func (f *fakeWorkflowExecution) SetExecutionData(executionData temporal_client.ExecutionData) {
	f.executionData = executionData
}

func (f *fakeWorkflowExecution) GetExecutionData() temporal_client.ExecutionData {
	return f.executionData
}

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

type fakeExerciseRepository struct {
	latest    *models.Exercise
	latestErr error
	saved     *models.Exercise
	saveErr   error
	updated   *models.Exercise
	updateErr error
	nextID    uint64
}

func (f *fakeExerciseRepository) Save(_ context.Context, entity *models.Exercise) (*models.Exercise, error) {
	if f.saveErr != nil {
		return nil, f.saveErr
	}
	copied := *entity
	if copied.ID == 0 {
		if f.nextID == 0 {
			f.nextID = 21
		}
		copied.ID = f.nextID
	}
	f.saved = &copied
	return f.saved, nil
}

func (f *fakeExerciseRepository) FindByID(context.Context, uint64) (*models.Exercise, error) {
	return nil, sql.ErrNoRows
}

func (f *fakeExerciseRepository) FindAll(context.Context, uint64, uint64) ([]*models.Exercise, error) {
	return nil, nil
}

func (f *fakeExerciseRepository) Update(_ context.Context, entity *models.Exercise) (*models.Exercise, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	copied := *entity
	f.updated = &copied
	return f.updated, nil
}

func (f *fakeExerciseRepository) DeleteByID(context.Context, uint64) error { return nil }

func (f *fakeExerciseRepository) FindLatestByUserID(context.Context, uint64) (*models.Exercise, error) {
	if f.latestErr != nil {
		return nil, f.latestErr
	}
	if f.latest == nil {
		return nil, sql.ErrNoRows
	}
	return f.latest, nil
}

func (f *fakeExerciseRepository) FindAllByUserID(context.Context, uint64, int64, int64) ([]*models.Exercise, error) {
	return nil, nil
}

func (f *fakeExerciseRepository) CountByUserID(context.Context, uint64) (int64, error) { return 0, nil }

func (f *fakeExerciseRepository) UpdateStatusIfCurrent(context.Context, uint64, string, string, string) (bool, error) {
	return true, nil
}

type fakeUserRepository struct {
	user    *models.User
	findErr error
	saveErr error
	saved   *models.User
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
	if f.findErr != nil {
		return nil, f.findErr
	}
	if f.user == nil {
		return nil, sql.ErrNoRows
	}
	return f.user, nil
}

func (f *fakeUserRepository) FindAll(context.Context, uint64, uint64) ([]*models.User, error) {
	return nil, nil
}

func (f *fakeUserRepository) Update(_ context.Context, entity *models.User) (*models.User, error) {
	copied := *entity
	return &copied, nil
}

func (f *fakeUserRepository) FindByOnboardingID(context.Context, uint64) (*models.User, error) {
	return nil, nil
	if f.findErr != nil {
		return nil, f.findErr
	}
	if f.user == nil {
		return nil, sql.ErrNoRows
	}
	return f.user, nil
}

func (f *fakeUserRepository) DeleteByID(context.Context, uint64) error { return nil }

type fakeExerciseCache struct {
	setErr         error
	setByKeyErr    error
	deleteErr      error
	deleteByKeyErr error
	setIDs         []uint64
	setKeys        []string
	deletedIDs     []uint64
	deletedKeys    []string
}

func (f *fakeExerciseCache) SetExercise(_ context.Context, id uint64, _ *responses.GenerateExerciseResponse) error {
	f.setIDs = append(f.setIDs, id)
	return f.setErr
}

func (f *fakeExerciseCache) GetExercise(context.Context, uint64) (*responses.GenerateExerciseResponse, error) {
	return nil, nil
}

func (f *fakeExerciseCache) SetExerciseByIdempotencyKey(_ context.Context, key string, _ *responses.GenerateExerciseResponse) error {
	f.setKeys = append(f.setKeys, key)
	return f.setByKeyErr
}

func (f *fakeExerciseCache) GetExerciseByIdempotencyKey(context.Context, string) (*responses.GenerateExerciseResponse, error) {
	return nil, nil
}

func (f *fakeExerciseCache) DeleteExercise(_ context.Context, id uint64) error {
	f.deletedIDs = append(f.deletedIDs, id)
	return f.deleteErr
}

func (f *fakeExerciseCache) DeleteExerciseByIdempotencyKey(_ context.Context, key string) error {
	f.deletedKeys = append(f.deletedKeys, key)
	return f.deleteByKeyErr
}

func (f *fakeExerciseCache) SetSubmitExercise(context.Context, uint64, *responses.SubmitExerciseResponse) error {
	return nil
}

func (f *fakeExerciseCache) GetSubmitExercise(context.Context, uint64) (*responses.SubmitExerciseResponse, error) {
	return nil, nil
}

func (f *fakeExerciseCache) DeleteSubmitExercise(context.Context, uint64) error { return nil }

type fakeDeepSeek struct {
	result string
	err    error
	prompt string
}

func (f *fakeDeepSeek) ChatGenerate(_ context.Context, prompt string) (string, error) {
	f.prompt = prompt
	return f.result, f.err
}
