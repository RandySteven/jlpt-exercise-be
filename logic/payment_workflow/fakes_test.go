package payment_workflow

import (
	"context"
	"database/sql"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/schemas"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
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
	args       []any

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

	executionData temporal_client.ExecutionData
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

func (f *fakeWorkflowExecution) StartWorkflow(_ context.Context, opts temporal_client.StartWorkflowOptions, _ interface{}, args ...interface{}) (client.WorkflowRun, error) {
	f.started = true
	f.startOpts = opts
	f.args = args
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

type fakePaymentRepository struct {
	byUserID        []*models.Payment
	userIDReads     int
	userIDErr       error
	byReference     *models.Payment
	referenceReads  int
	referenceErr    error
	bySubscription  *models.Payment
	subscriptionErr error
	byID            *models.Payment
	findErr         error
	saved           *models.Payment
	saveErr         error
	updated         *models.Payment
	updateErr       error
	nextID          uint64
}

func (f *fakePaymentRepository) Save(_ context.Context, entity *models.Payment) (*models.Payment, error) {
	if f.saveErr != nil {
		return nil, f.saveErr
	}
	copied := *entity
	if copied.ID == 0 {
		if f.nextID == 0 {
			f.nextID = 41
		}
		copied.ID = f.nextID
	}
	f.saved = &copied
	return f.saved, nil
}

func (f *fakePaymentRepository) FindByID(context.Context, uint64) (*models.Payment, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.byID, nil
}

func (f *fakePaymentRepository) FindAll(context.Context, uint64, uint64) ([]*models.Payment, error) {
	return nil, nil
}

func (f *fakePaymentRepository) Update(_ context.Context, entity *models.Payment) (*models.Payment, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	copied := *entity
	f.updated = &copied
	return f.updated, nil
}

func (f *fakePaymentRepository) DeleteByID(context.Context, uint64) error { return nil }

func (f *fakePaymentRepository) FindByReferenceID(context.Context, string) (*models.Payment, error) {
	f.referenceReads++
	if f.referenceReads == 1 && f.referenceErr != nil {
		return nil, f.referenceErr
	}
	if f.byReference == nil && f.referenceErr != nil {
		return nil, f.referenceErr
	}
	return f.byReference, nil
}

func (f *fakePaymentRepository) FindBySubscriptionID(context.Context, uint64) (*models.Payment, error) {
	if f.subscriptionErr != nil {
		return nil, f.subscriptionErr
	}
	if f.bySubscription == nil {
		return nil, sql.ErrNoRows
	}
	return f.bySubscription, nil
}

func (f *fakePaymentRepository) FindByUserID(context.Context, uint64) ([]*models.Payment, error) {
	f.userIDReads++
	if f.userIDErr != nil {
		return nil, f.userIDErr
	}
	return f.byUserID, nil
}

type fakeSubscriptionRepository struct {
	current    *models.Subscription
	currentErr error
	byPayment  *models.Subscription
	paymentErr error
	saved      *models.Subscription
	saveErr    error
	updated    *models.Subscription
	updateErr  error
	nextID     uint64
}

func (f *fakeSubscriptionRepository) Save(_ context.Context, entity *models.Subscription) (*models.Subscription, error) {
	if f.saveErr != nil {
		return nil, f.saveErr
	}
	copied := *entity
	if copied.ID == 0 {
		if f.nextID == 0 {
			f.nextID = 51
		}
		copied.ID = f.nextID
	}
	f.saved = &copied
	return f.saved, nil
}

func (f *fakeSubscriptionRepository) FindByID(context.Context, uint64) (*models.Subscription, error) {
	return nil, nil
}

func (f *fakeSubscriptionRepository) FindAll(context.Context, uint64, uint64) ([]*models.Subscription, error) {
	return nil, nil
}

func (f *fakeSubscriptionRepository) Update(_ context.Context, entity *models.Subscription) (*models.Subscription, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	copied := *entity
	f.updated = &copied
	return f.updated, nil
}

func (f *fakeSubscriptionRepository) DeleteByID(context.Context, uint64) error { return nil }

func (f *fakeSubscriptionRepository) FindByReferenceID(context.Context, string) (*models.Subscription, error) {
	return nil, nil
}

func (f *fakeSubscriptionRepository) FindByPaymentID(context.Context, uint64) (*models.Subscription, error) {
	if f.paymentErr != nil {
		return nil, f.paymentErr
	}
	return f.byPayment, nil
}

func (f *fakeSubscriptionRepository) FindByUserID(context.Context, uint64) ([]*models.Subscription, error) {
	return nil, nil
}

func (f *fakeSubscriptionRepository) FindCurrentByUserID(context.Context, uint64) (*models.Subscription, error) {
	if f.currentErr != nil {
		return nil, f.currentErr
	}
	return f.current, nil
}

type fakePlanRepository struct {
	plan    *models.SubscriptionPlan
	findErr error
}

func (f *fakePlanRepository) Save(context.Context, *models.SubscriptionPlan) (*models.SubscriptionPlan, error) {
	return nil, nil
}

func (f *fakePlanRepository) FindByID(context.Context, uint64) (*models.SubscriptionPlan, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.plan, nil
}

func (f *fakePlanRepository) FindAll(context.Context, uint64, uint64) ([]*models.SubscriptionPlan, error) {
	return nil, nil
}

func (f *fakePlanRepository) Update(context.Context, *models.SubscriptionPlan) (*models.SubscriptionPlan, error) {
	return nil, nil
}

func (f *fakePlanRepository) DeleteByID(context.Context, uint64) error { return nil }

type fakeUserRepository struct {
	user    *models.User
	findErr error
}

func (f *fakeUserRepository) Save(context.Context, *models.User) (*models.User, error) {
	return nil, nil
}

func (f *fakeUserRepository) FindByID(context.Context, uint64) (*models.User, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.user, nil
}

func (f *fakeUserRepository) FindAll(context.Context, uint64, uint64) ([]*models.User, error) {
	return nil, nil
}

func (f *fakeUserRepository) Update(context.Context, *models.User) (*models.User, error) {
	return nil, nil
}

func (f *fakeUserRepository) DeleteByID(context.Context, uint64) error { return nil }

type fakeOnboardingRepository struct {
	onboarding *models.Onboarding
	findErr    error
}

func (f *fakeOnboardingRepository) Save(context.Context, *models.Onboarding) (*models.Onboarding, error) {
	return nil, nil
}

func (f *fakeOnboardingRepository) Update(context.Context, *models.Onboarding) (*models.Onboarding, error) {
	return nil, nil
}

func (f *fakeOnboardingRepository) FindByEmail(context.Context, string) (*models.Onboarding, error) {
	return nil, nil
}

func (f *fakeOnboardingRepository) FindByID(context.Context, uint64) (*models.Onboarding, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.onboarding, nil
}

type fakeMidtrans struct {
	response *midtrans_client.TransactionResponse
	err      error
	got      *midtrans_client.TransactionRequest
}

func (f *fakeMidtrans) CreateTransaction(_ context.Context, request *midtrans_client.TransactionRequest) (*midtrans_client.TransactionResponse, error) {
	copied := *request
	f.got = &copied
	return f.response, f.err
}

func (f *fakeMidtrans) Charge(context.Context, *midtrans_client.ChargeRequest) (*midtrans_client.ChargeResponse, error) {
	return nil, nil
}

func (f *fakeMidtrans) Capture(context.Context, *midtrans_client.CaptureRequest) (*midtrans_client.CaptureResponse, error) {
	return nil, nil
}

func (f *fakeMidtrans) Cancel(context.Context, *midtrans_client.CancelRequest) (*midtrans_client.CancelResponse, error) {
	return nil, nil
}

type fakePaymentTopic struct {
	schema *schemas.Payment
	err    error
}

func (f *fakePaymentTopic) WriteMessage(_ context.Context, schema *schemas.Payment) error {
	f.schema = schema
	return f.err
}

type fakeSMTP struct {
	email   string
	subject string
	body    string
	err     error
}

func (f *fakeSMTP) SendEmail(email string, subject string, body string) error {
	f.email = email
	f.subject = subject
	f.body = body
	return f.err
}

func (f *fakeSMTP) SendHTMLTemplate(email string, subject string, templateName string, _ any) error {
	f.email = email
	f.subject = subject
	f.body = templateName
	return f.err
}
