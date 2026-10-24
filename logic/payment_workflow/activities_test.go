package payment_workflow

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/schemas"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
	midtransmocks "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans/mocks"
	smtpmocks "github.com/RandySteven/jlpt-exercise-be/pkg/smtp/mocks"
	repomocks "github.com/RandySteven/jlpt-exercise-be/repositories/mocks"
	topicmocks "github.com/RandySteven/jlpt-exercise-be/topics/mocks"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"
)

func TestCheckPlan(t *testing.T) {
	workflow, _ := readyWorkflow(t)

	_, err := workflow.checkPlan(context.Background(), nil)
	requireInputError(t, err, "subscription request is missing")
	_, err = workflow.checkPlan(context.Background(), &ExecutionData{})
	requireInputError(t, err, "subscription request is missing")
	_, err = workflow.checkPlan(context.Background(), &ExecutionData{Request: &requests.CreateSubscriptionRequest{}, UserID: 4})
	requireInputError(t, err, "subscription request is incomplete")

	users := repomocks.NewMockUserRepository(t)
	users.EXPECT().FindByID(mock.Anything, uint64(4)).Return(nil, sql.ErrNoRows)
	workflow.userRepository = users
	_, err = workflow.checkPlan(context.Background(), readyExecution())
	requireInputError(t, err, "user does not exist")

	users = repomocks.NewMockUserRepository(t)
	users.EXPECT().FindByID(mock.Anything, uint64(4)).Return(nil, errors.New("db down"))
	workflow.userRepository = users
	_, err = workflow.checkPlan(context.Background(), readyExecution())
	if err == nil || err.Error() != "db down" {
		t.Fatalf("user lookup: %v", err)
	}

	workflow, _ = readyWorkflow(t)
	onboarding := repomocks.NewMockOnboardingRepository(t)
	onboarding.EXPECT().FindByID(mock.Anything, uint64(6)).Return(nil, sql.ErrNoRows)
	workflow.onboardingRepository = onboarding
	_, err = workflow.checkPlan(context.Background(), readyExecution())
	requireInputError(t, err, "onboarding does not exist")

	workflow, _ = readyWorkflow(t)
	onboarding = repomocks.NewMockOnboardingRepository(t)
	onboarding.EXPECT().FindByID(mock.Anything, uint64(6)).Return(&models.Onboarding{Email: ""}, nil)
	workflow.onboardingRepository = onboarding
	_, err = workflow.checkPlan(context.Background(), readyExecution())
	requireInputError(t, err, "customer email is missing")

	workflow, _ = readyWorkflow(t)
	plans := repomocks.NewMockSubscriptionPlanRepository(t)
	plans.EXPECT().FindByID(mock.Anything, uint64(3)).Return(nil, sql.ErrNoRows)
	workflow.planRepository = plans
	_, err = workflow.checkPlan(context.Background(), readyExecution())
	requireInputError(t, err, "plan does not exist")

	workflow, _ = readyWorkflow(t)
	plans = repomocks.NewMockSubscriptionPlanRepository(t)
	plans.EXPECT().FindByID(mock.Anything, uint64(3)).Return(&models.SubscriptionPlan{ID: 3, Price: 0, Currency: "IDR"}, nil)
	workflow.planRepository = plans
	_, err = workflow.checkPlan(context.Background(), readyExecution())
	requireInputError(t, err, "plan price is invalid")

	workflow, _ = readyWorkflow(t)
	subscriptions := repomocks.NewMockSubscriptionRepository(t)
	subscriptions.EXPECT().FindCurrentByUserID(mock.Anything, uint64(4)).Return(&models.Subscription{ID: 9, Status: subscriptionStatusActive}, nil)
	workflow.subscriptionRepository = subscriptions
	got, err := workflow.checkPlan(context.Background(), readyExecution())
	if err != nil {
		t.Fatal(err)
	}
	if got.CurrActivity != planExistActivityName || got.Response.SubscriptionID != 9 || got.CustomerEmail != "ada@example.com" {
		t.Fatalf("existing subscription = %#v", got)
	}

	workflow, _ = readyWorkflow(t)
	subscriptions = repomocks.NewMockSubscriptionRepository(t)
	subscriptions.EXPECT().FindCurrentByUserID(mock.Anything, uint64(4)).Return(&models.Subscription{ID: 11, Status: subscriptionStatusPending}, nil)
	workflow.subscriptionRepository = subscriptions
	got, err = workflow.checkPlan(context.Background(), readyExecution())
	if err != nil || got.CurrActivity != persistPaymentActivityName || got.Response.SubscriptionID != 11 {
		t.Fatalf("pending subscription = %#v err %v", got, err)
	}

	workflow, _ = readyWorkflow(t)
	got, err = workflow.checkPlan(context.Background(), readyExecution())
	if err != nil {
		t.Fatal(err)
	}
	if got.CurrActivity != createSubscriptionActivityName || got.Plan.ID != 3 || got.CustomerName != "Ada" {
		t.Fatalf("new subscription = %#v", got)
	}
}

func TestPersistPaymentAndSubscription(t *testing.T) {
	workflow, deps := readyWorkflow(t)
	_, err := workflow.persistPayment(context.Background(), &ExecutionData{})
	requireInputError(t, err, "payment input is missing")
	_, err = workflow.createSubscription(context.Background(), &ExecutionData{})
	requireInputError(t, err, "subscription input is missing")

	data := readyExecution()
	data.Plan = monthPlan()
	got, err := workflow.createSubscription(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	subscription := deps.subscriptions.saved
	end, endErr := subscriptionEnd(subscription.StartDate, data.Plan)
	if endErr != nil || !subscription.EndDate.Equal(end) || subscription.Status != subscriptionStatusPending || subscription.ReferenceID != "key" || got.CurrActivity != persistPaymentActivityName || got.Response.SubscriptionID != subscription.ID {
		t.Fatalf("subscription = %#v activity %s", subscription, got.CurrActivity)
	}

	got, err = workflow.persistPayment(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	saved := deps.payments.saved
	if got.CurrActivity != midtransCallActivityName || saved.ReferenceID != "key" || saved.UserID != 4 || saved.SubscriptionID != subscription.ID || saved.Amount != 15000 || saved.Currency != "IDR" || saved.Status != paymentStatusPending {
		t.Fatalf("saved payment = %#v activity %s", saved, got.CurrActivity)
	}
	if deps.execution.signalName != sgMidtransCall {
		t.Fatalf("midtrans signal = %s", deps.execution.signalName)
	}

	payments, _ := paymentRepository(t, nil, &models.Payment{ID: 8, UserID: 4, ReferenceID: "key"}, nil, nil)
	workflow.paymentRepository = payments
	got, err = workflow.persistPayment(context.Background(), data)
	if err != nil || got.Payment.ID != 8 || got.CurrActivity != midtransCallActivityName {
		t.Fatalf("reused payment = %#v err %v", got, err)
	}

	payments, _ = paymentRepository(t, nil, &models.Payment{ID: 8, UserID: 99, ReferenceID: "key"}, nil, nil)
	workflow.paymentRepository = payments
	_, err = workflow.persistPayment(context.Background(), data)
	requireInputError(t, err, "payment reference belongs to another user")

	existing := &models.Payment{ID: 12, UserID: 4, ReferenceID: "key"}
	payments, _ = paymentRepository(t, sql.ErrNoRows, existing, &pgconn.PgError{Code: "23505"}, nil)
	workflow.paymentRepository = payments
	got, err = workflow.persistPayment(context.Background(), data)
	if err != nil || got.Payment.ID != 12 || got.CurrActivity != midtransCallActivityName {
		t.Fatalf("unique payment = %#v err %v", got, err)
	}

	data.Plan = &models.SubscriptionPlan{ID: 3, Price: 15000, Currency: "IDR", Interval: "decade", IntervalCount: 1}
	data.Subscription = nil
	_, err = workflow.createSubscription(context.Background(), data)
	requireInputError(t, err, `plan interval "decade" is invalid`)
}

func TestMidtransCall(t *testing.T) {
	workflow, _ := readyWorkflow(t)
	_, err := workflow.midtransCall(context.Background(), &ExecutionData{})
	requireInputError(t, err, "midtrans input is missing")

	workflow, deps := readyWorkflow(t)
	midtrans, request := midtransClient(t, snapResponse(), nil)
	workflow.midtransClient = midtrans
	data := chargedExecution()
	got, err := workflow.midtransCall(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	if got.CurrActivity != updatePaymentStatusActivityName || got.Response.SubscriptionID != 51 || got.Response.RedirectURL != "https://pay.example/snap" || got.MidtransTransaction.RedirectURL != "https://pay.example/snap" {
		t.Fatalf("midtrans result = %#v", got)
	}
	gotRequest := request.got
	if gotRequest.TransactionDetails.OrderID != "key" || gotRequest.TransactionDetails.GrossAmount != 15000 || gotRequest.CustomerDetails.Email != "ada@example.com" || gotRequest.CustomField1 != "51" {
		t.Fatalf("midtrans request = %#v", gotRequest)
	}
	if !strings.Contains(string(deps.payments.updated.Metadata), "https://pay.example/snap") {
		t.Fatalf("metadata = %s", deps.payments.updated.Metadata)
	}

	workflow, _ = readyWorkflow(t)
	midtrans, _ = midtransClient(t, nil, errors.New("status 400"))
	workflow.midtransClient = midtrans
	got, err = workflow.midtransCall(context.Background(), chargedExecution())
	if err != nil || got.CurrActivity != paymentFailedActivityName {
		t.Fatalf("rejected = %#v err %v", got, err)
	}

	midtrans, _ = midtransClient(t, nil, errors.New("timeout"))
	workflow.midtransClient = midtrans
	_, err = workflow.midtransCall(context.Background(), chargedExecution())
	if err == nil || err.Error() != "timeout" {
		t.Fatalf("retryable midtrans error = %v", err)
	}

	workflow, _ = readyWorkflow(t)
	payments, _ := paymentRepository(t, nil, nil, nil, &models.Payment{
		ID: 41, ReferenceID: "key", Amount: 15000, UserID: 4,
		Metadata: []byte(`{"token":"snap-token","redirect_url":"https://pay.example/again"}`),
	})
	workflow.paymentRepository = payments
	var called bool
	midtrans = midtransmocks.NewMockMidtransClient(t)
	midtrans.EXPECT().CreateTransaction(mock.Anything, mock.Anything).Run(func(context.Context, *midtrans_client.TransactionRequest) {
		called = true
	}).Return(nil, nil).Maybe()
	workflow.midtransClient = midtrans
	got, err = workflow.midtransCall(context.Background(), chargedExecution())
	if err != nil || called || got.MidtransTransaction.Token != "snap-token" || got.CurrActivity != updatePaymentStatusActivityName {
		t.Fatalf("replay = %#v called %v err %v", got.MidtransTransaction, called, err)
	}
}

func TestConfirmPublishAndEmail(t *testing.T) {
	workflow, deps := readyWorkflow(t)
	_, err := workflow.confirmPayment(context.Background(), &ExecutionData{})
	requireInputError(t, err, "payment confirmation is missing")

	data := chargedExecution()
	data.ConfirmRequest = &requests.PaymentConfirmationRequest{IdempotencyKey: "key", PlanID: 3, Status: "capture"}
	before := time.Now().UTC()
	got, err := workflow.confirmPayment(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	payment := deps.payments.updated
	subscription := deps.subscriptions.updated
	if payment.Status != paymentStatusPaid || subscription.Status != subscriptionStatusActive || got.CurrActivity != publishPaymentActivityName {
		t.Fatalf("confirmed payment %#v subscription %#v", payment, subscription)
	}
	if subscription.StartDate.Before(before) {
		t.Fatalf("start date was not reset: %s", subscription.StartDate)
	}

	got, err = workflow.publishPayment(context.Background(), got)
	if err != nil || got.CurrActivity != sendSubscriptionEmailActivityName || deps.topic.ID != "41" || deps.topic.Status != paymentStatusPaid {
		t.Fatalf("publish = %#v schema %#v err %v", got, deps.topic, err)
	}

	got, err = workflow.sendSubscriptionEmail(context.Background(), got)
	if err != nil || deps.smtpEmail != "ada@example.com" || !strings.Contains(deps.smtpBody, "N3") {
		t.Fatalf("email = %s body %s err %v", deps.smtpEmail, deps.smtpBody, err)
	}

	smtp := smtpmocks.NewMockSmtpClient(t)
	smtp.EXPECT().SendEmail(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("smtp down"))
	workflow.smtpClient = smtp
	if _, err = workflow.sendSubscriptionEmail(context.Background(), got); err != nil {
		t.Fatalf("email failure should not fail the paid subscription: %v", err)
	}

	_, err = workflow.paymentFailed(context.Background(), chargedExecution())
	requireRejected(t, err)
	if deps.payments.updated.Status != paymentStatusFailed || deps.subscriptions.updated.Status != subscriptionStatusFailed {
		t.Fatalf("failed payment %#v subscription %#v", deps.payments.updated, deps.subscriptions.updated)
	}
}

type paymentDeps struct {
	execution     *workflowDouble
	payments      *modelCapture[*models.Payment]
	subscriptions *modelCapture[*models.Subscription]
	topic         *schemas.Payment
	smtpEmail     string
	smtpBody      string
}

type modelCapture[T any] struct {
	saved   T
	updated T
}

func readyWorkflow(t *testing.T) (*paymentWorkflow, *paymentDeps) {
	t.Helper()
	deps := &paymentDeps{execution: newWorkflowDouble(t)}
	deps.execution.allowSignal(nil)
	payments, paymentCapture := paymentRepositoryState(t, sql.ErrNoRows, nil, nil, nil)
	deps.payments = paymentCapture
	subscriptions, subscriptionCapture := subscriptionRepository(t, nil, nil)
	deps.subscriptions = subscriptionCapture

	users := repomocks.NewMockUserRepository(t)
	users.EXPECT().FindByID(mock.Anything, mock.Anything).Return(&models.User{ID: 4, OnboardingID: 6}, nil).Maybe()
	onboarding := repomocks.NewMockOnboardingRepository(t)
	onboarding.EXPECT().FindByID(mock.Anything, mock.Anything).Return(&models.Onboarding{ID: 6, Name: "Ada", Email: "ada@example.com"}, nil).Maybe()
	plans := repomocks.NewMockSubscriptionPlanRepository(t)
	plans.EXPECT().FindByID(mock.Anything, mock.Anything).Return(monthPlan(), nil).Maybe()
	midtrans := midtransmocks.NewMockMidtransClient(t)
	midtrans.EXPECT().CreateTransaction(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	topic := topicmocks.NewMockPaymentTopic(t)
	topic.EXPECT().WriteMessage(mock.Anything, mock.Anything).Run(func(_ context.Context, schema *schemas.Payment) {
		copied := *schema
		deps.topic = &copied
	}).Return(nil).Maybe()
	smtp := smtpmocks.NewMockSmtpClient(t)
	smtp.EXPECT().SendEmail(mock.Anything, mock.Anything, mock.Anything).Run(func(email string, _ string, body string) {
		deps.smtpEmail = email
		deps.smtpBody = body
	}).Return(nil).Maybe()

	return &paymentWorkflow{
		workflow:               deps.execution.exec,
		paymentRepository:      payments,
		subscriptionRepository: subscriptions,
		planRepository:         plans,
		userRepository:         users,
		onboardingRepository:   onboarding,
		midtransClient:         midtrans,
		paymentTopic:           topic,
		smtpClient:             smtp,
	}, deps
}

func paymentRepository(t *testing.T, referenceErr error, byReference *models.Payment, saveErr error, byID *models.Payment) (*repomocks.MockPaymentRepository, *modelCapture[*models.Payment]) {
	t.Helper()
	return paymentRepositoryState(t, referenceErr, byReference, saveErr, byID)
}

func paymentRepositoryState(t *testing.T, referenceErr error, byReference *models.Payment, saveErr error, byID *models.Payment) (*repomocks.MockPaymentRepository, *modelCapture[*models.Payment]) {
	t.Helper()
	repo := repomocks.NewMockPaymentRepository(t)
	captured := &modelCapture[*models.Payment]{}
	reads := 0
	repo.EXPECT().FindByReferenceID(mock.Anything, mock.Anything).RunAndReturn(func(context.Context, string) (*models.Payment, error) {
		reads++
		if reads == 1 && referenceErr != nil {
			return nil, referenceErr
		}
		if byReference == nil && referenceErr != nil {
			return nil, referenceErr
		}
		return byReference, nil
	}).Maybe()
	repo.EXPECT().FindBySubscriptionID(mock.Anything, mock.Anything).Return(nil, sql.ErrNoRows).Maybe()
	repo.EXPECT().FindByID(mock.Anything, mock.Anything).Return(byID, nil).Maybe()
	repo.EXPECT().Save(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, entity *models.Payment) (*models.Payment, error) {
		if saveErr != nil {
			return nil, saveErr
		}
		copied := *entity
		if copied.ID == 0 {
			copied.ID = 41
		}
		captured.saved = &copied
		return captured.saved, nil
	}).Maybe()
	repo.EXPECT().Update(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, entity *models.Payment) (*models.Payment, error) {
		copied := *entity
		captured.updated = &copied
		return captured.updated, nil
	}).Maybe()
	return repo, captured
}

func subscriptionRepository(t *testing.T, current *models.Subscription, currentErr error) (*repomocks.MockSubscriptionRepository, *modelCapture[*models.Subscription]) {
	t.Helper()
	repo := repomocks.NewMockSubscriptionRepository(t)
	captured := &modelCapture[*models.Subscription]{}
	repo.EXPECT().FindCurrentByUserID(mock.Anything, mock.Anything).Return(current, currentErr).Maybe()
	repo.EXPECT().Save(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, entity *models.Subscription) (*models.Subscription, error) {
		copied := *entity
		if copied.ID == 0 {
			copied.ID = 51
		}
		captured.saved = &copied
		return captured.saved, nil
	}).Maybe()
	repo.EXPECT().Update(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, entity *models.Subscription) (*models.Subscription, error) {
		copied := *entity
		captured.updated = &copied
		return captured.updated, nil
	}).Maybe()
	return repo, captured
}

func midtransClient(t *testing.T, response *midtrans_client.TransactionResponse, callErr error) (*midtransmocks.MockMidtransClient, *struct {
	got *midtrans_client.TransactionRequest
}) {
	t.Helper()
	client := midtransmocks.NewMockMidtransClient(t)
	captured := &struct {
		got *midtrans_client.TransactionRequest
	}{}
	client.EXPECT().CreateTransaction(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, request *midtrans_client.TransactionRequest) (*midtrans_client.TransactionResponse, error) {
		if request != nil {
			copied := *request
			captured.got = &copied
		}
		return response, callErr
	})
	return client, captured
}

func readyExecution() *ExecutionData {
	return &ExecutionData{
		UserID: 4,
		Request: &requests.CreateSubscriptionRequest{
			IdempotencyKey: "key",
			PlanID:         3,
		},
	}
}

func monthPlan() *models.SubscriptionPlan {
	return &models.SubscriptionPlan{ID: 3, Name: "N3", Price: 15000, Currency: "IDR", Interval: "month", IntervalCount: 1}
}

func chargedExecution() *ExecutionData {
	data := readyExecution()
	data.Plan = monthPlan()
	data.CustomerEmail = "ada@example.com"
	data.CustomerName = "Ada"
	data.Payment = &models.Payment{ID: 41, ReferenceID: "key", UserID: 4, Amount: 15000, Currency: "IDR", Status: paymentStatusPending}
	data.Subscription = &models.Subscription{ID: 51, UserID: 4, PlanID: 3, Status: subscriptionStatusPending}
	return data
}

func snapResponse() *midtrans_client.TransactionResponse {
	return &midtrans_client.TransactionResponse{Token: "snap-token", RedirectURL: "https://pay.example/snap"}
}

func TestSubscriptionEnd(t *testing.T) {
	start := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	end, err := subscriptionEnd(start, &models.SubscriptionPlan{Interval: "month", IntervalCount: 1})
	if err != nil || !end.Equal(time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("month end = %s err %v", end, err)
	}
	end, err = subscriptionEnd(start, &models.SubscriptionPlan{Interval: "week", IntervalCount: 2})
	if err != nil || !end.Equal(start.AddDate(0, 0, 14)) {
		t.Fatalf("week end = %s err %v", end, err)
	}
	if _, err = subscriptionEnd(start, &models.SubscriptionPlan{Interval: "month", IntervalCount: 0}); err == nil {
		t.Fatal("expected invalid interval count")
	}
}

func requireInputError(t *testing.T, err error, message string) {
	t.Helper()
	var appErr *temporal.ApplicationError
	if !errors.As(err, &appErr) || appErr.Type() != paymentInputErrorType || !appErr.NonRetryable() || appErr.Message() != message {
		t.Fatalf("got %v, want non-retryable %s %q", err, paymentInputErrorType, message)
	}
}

func requireRejected(t *testing.T, err error) {
	t.Helper()
	var appErr *temporal.ApplicationError
	if !errors.As(err, &appErr) || appErr.Type() != paymentRejectedErrorType || !appErr.NonRetryable() {
		t.Fatalf("got %v, want non-retryable %s", err, paymentRejectedErrorType)
	}
}

func requireCustom(t *testing.T, err error, kind apperror.ErrType) *apperror.CustomError {
	t.Helper()
	custom, ok := err.(*apperror.CustomError)
	if !ok || custom == nil || custom.ErrType != kind {
		t.Fatalf("got %#v, want custom error %v", err, kind)
	}
	return custom
}
