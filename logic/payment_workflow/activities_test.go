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
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
	"github.com/jackc/pgx/v5/pgconn"
	"go.temporal.io/sdk/temporal"
)

func TestCheckPlan(t *testing.T) {
	workflow := readyWorkflow()

	_, err := workflow.checkPlan(context.Background(), nil)
	requireInputError(t, err, "subscription request is missing")
	_, err = workflow.checkPlan(context.Background(), &ExecutionData{})
	requireInputError(t, err, "subscription request is missing")
	_, err = workflow.checkPlan(context.Background(), &ExecutionData{Request: &requests.CreateSubscriptionRequest{}, UserID: 4})
	requireInputError(t, err, "subscription request is incomplete")

	workflow.userRepository = &fakeUserRepository{findErr: sql.ErrNoRows}
	_, err = workflow.checkPlan(context.Background(), readyExecution())
	requireInputError(t, err, "user does not exist")

	workflow.userRepository = &fakeUserRepository{findErr: errors.New("db down")}
	_, err = workflow.checkPlan(context.Background(), readyExecution())
	if err == nil || err.Error() != "db down" {
		t.Fatalf("user lookup: %v", err)
	}

	workflow = readyWorkflow()
	workflow.onboardingRepository = &fakeOnboardingRepository{findErr: sql.ErrNoRows}
	_, err = workflow.checkPlan(context.Background(), readyExecution())
	requireInputError(t, err, "onboarding does not exist")

	workflow = readyWorkflow()
	workflow.onboardingRepository = &fakeOnboardingRepository{onboarding: &models.Onboarding{Email: ""}}
	_, err = workflow.checkPlan(context.Background(), readyExecution())
	requireInputError(t, err, "customer email is missing")

	workflow = readyWorkflow()
	workflow.planRepository = &fakePlanRepository{findErr: sql.ErrNoRows}
	_, err = workflow.checkPlan(context.Background(), readyExecution())
	requireInputError(t, err, "plan does not exist")

	workflow = readyWorkflow()
	workflow.planRepository = &fakePlanRepository{plan: &models.SubscriptionPlan{ID: 3, Price: 0, Currency: "IDR"}}
	_, err = workflow.checkPlan(context.Background(), readyExecution())
	requireInputError(t, err, "plan price is invalid")

	workflow = readyWorkflow()
	workflow.subscriptionRepository = &fakeSubscriptionRepository{current: &models.Subscription{ID: 9, Status: subscriptionStatusActive}}
	got, err := workflow.checkPlan(context.Background(), readyExecution())
	if err != nil {
		t.Fatal(err)
	}
	if got.CurrActivity != planExistActivityName || got.Response.SubscriptionID != 9 || got.CustomerEmail != "ada@example.com" {
		t.Fatalf("existing subscription = %#v", got)
	}

	workflow = readyWorkflow()
	workflow.subscriptionRepository = &fakeSubscriptionRepository{current: &models.Subscription{ID: 11, Status: subscriptionStatusPending}}
	got, err = workflow.checkPlan(context.Background(), readyExecution())
	if err != nil || got.CurrActivity != persistPaymentActivityName || got.Response.SubscriptionID != 11 {
		t.Fatalf("pending subscription = %#v err %v", got, err)
	}

	workflow = readyWorkflow()
	got, err = workflow.checkPlan(context.Background(), readyExecution())
	if err != nil {
		t.Fatal(err)
	}
	if got.CurrActivity != createSubscriptionActivityName || got.Plan.ID != 3 || got.CustomerName != "Ada" {
		t.Fatalf("new subscription = %#v", got)
	}
}

func TestPersistPaymentAndSubscription(t *testing.T) {
	workflow := readyWorkflow()
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
	subscription := workflow.subscriptionRepository.(*fakeSubscriptionRepository).saved
	end, endErr := subscriptionEnd(subscription.StartDate, data.Plan)
	if endErr != nil || !subscription.EndDate.Equal(end) || subscription.Status != subscriptionStatusPending || subscription.ReferenceID != "key" || got.CurrActivity != persistPaymentActivityName || got.Response.SubscriptionID != subscription.ID {
		t.Fatalf("subscription = %#v activity %s", subscription, got.CurrActivity)
	}

	got, err = workflow.persistPayment(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	saved := workflow.paymentRepository.(*fakePaymentRepository).saved
	if got.CurrActivity != midtransCallActivityName || saved.ReferenceID != "key" || saved.UserID != 4 || saved.SubscriptionID != subscription.ID || saved.Amount != 15000 || saved.Currency != "IDR" || saved.Status != paymentStatusPending {
		t.Fatalf("saved payment = %#v activity %s", saved, got.CurrActivity)
	}
	if workflow.workflow.(*fakeWorkflowExecution).signalName != sgMidtransCall {
		t.Fatalf("midtrans signal = %s", workflow.workflow.(*fakeWorkflowExecution).signalName)
	}

	workflow.paymentRepository = &fakePaymentRepository{byReference: &models.Payment{ID: 8, UserID: 4, ReferenceID: "key"}}
	got, err = workflow.persistPayment(context.Background(), data)
	if err != nil || got.Payment.ID != 8 || got.CurrActivity != midtransCallActivityName {
		t.Fatalf("reused payment = %#v err %v", got, err)
	}

	workflow.paymentRepository = &fakePaymentRepository{byReference: &models.Payment{ID: 8, UserID: 99, ReferenceID: "key"}}
	_, err = workflow.persistPayment(context.Background(), data)
	requireInputError(t, err, "payment reference belongs to another user")

	existing := &models.Payment{ID: 12, UserID: 4, ReferenceID: "key"}
	workflow.paymentRepository = &fakePaymentRepository{
		referenceErr: sql.ErrNoRows,
		byReference:  existing,
		saveErr:      &pgconn.PgError{Code: "23505"},
	}
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
	workflow := readyWorkflow()
	_, err := workflow.midtransCall(context.Background(), &ExecutionData{})
	requireInputError(t, err, "midtrans input is missing")

	midtrans := &fakeMidtrans{response: snapResponse()}
	workflow.midtransClient = midtrans
	data := chargedExecution()
	got, err := workflow.midtransCall(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	if got.CurrActivity != updatePaymentStatusActivityName || got.Response.SubscriptionID != 51 || got.Response.RedirectURL != "https://pay.example/snap" || got.MidtransTransaction.RedirectURL != "https://pay.example/snap" {
		t.Fatalf("midtrans result = %#v", got)
	}
	if midtrans.got.TransactionDetails.OrderID != "key" || midtrans.got.TransactionDetails.GrossAmount != 15000 || midtrans.got.CustomerDetails.Email != "ada@example.com" || midtrans.got.CustomField1 != "51" {
		t.Fatalf("midtrans request = %#v", midtrans.got)
	}
	if !strings.Contains(string(workflow.paymentRepository.(*fakePaymentRepository).updated.Metadata), "https://pay.example/snap") {
		t.Fatalf("metadata = %s", workflow.paymentRepository.(*fakePaymentRepository).updated.Metadata)
	}

	workflow = readyWorkflow()
	workflow.midtransClient = &fakeMidtrans{err: errors.New("status 400")}
	got, err = workflow.midtransCall(context.Background(), chargedExecution())
	if err != nil || got.CurrActivity != paymentFailedActivityName {
		t.Fatalf("rejected = %#v err %v", got, err)
	}

	workflow.midtransClient = &fakeMidtrans{err: errors.New("timeout")}
	_, err = workflow.midtransCall(context.Background(), chargedExecution())
	if err == nil || err.Error() != "timeout" {
		t.Fatalf("retryable midtrans error = %v", err)
	}

	workflow = readyWorkflow()
	workflow.paymentRepository = &fakePaymentRepository{byID: &models.Payment{
		ID: 41, ReferenceID: "key", Amount: 15000, UserID: 4,
		Metadata: []byte(`{"token":"snap-token","redirect_url":"https://pay.example/again"}`),
	}}
	midtrans = &fakeMidtrans{}
	workflow.midtransClient = midtrans
	got, err = workflow.midtransCall(context.Background(), chargedExecution())
	if err != nil || midtrans.got != nil || got.MidtransTransaction.Token != "snap-token" || got.CurrActivity != updatePaymentStatusActivityName {
		t.Fatalf("replay = %#v called %#v err %v", got.MidtransTransaction, midtrans.got, err)
	}
}

func TestConfirmPublishAndEmail(t *testing.T) {
	workflow := readyWorkflow()
	_, err := workflow.confirmPayment(context.Background(), &ExecutionData{})
	requireInputError(t, err, "payment confirmation is missing")

	data := chargedExecution()
	before := time.Now().UTC()
	got, err := workflow.confirmPayment(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	payment := workflow.paymentRepository.(*fakePaymentRepository).updated
	subscription := workflow.subscriptionRepository.(*fakeSubscriptionRepository).updated
	if payment.Status != paymentStatusPaid || subscription.Status != subscriptionStatusActive || got.CurrActivity != publishPaymentActivityName {
		t.Fatalf("confirmed payment %#v subscription %#v", payment, subscription)
	}
	if subscription.StartDate.Before(before) {
		t.Fatalf("start date was not reset: %s", subscription.StartDate)
	}

	topic := &fakePaymentTopic{}
	workflow.paymentTopic = topic
	got, err = workflow.publishPayment(context.Background(), got)
	if err != nil || got.CurrActivity != sendSubscriptionEmailActivityName || topic.schema.ID != "41" || topic.schema.Status != paymentStatusPaid {
		t.Fatalf("publish = %#v schema %#v err %v", got, topic.schema, err)
	}

	smtp := &fakeSMTP{}
	workflow.smtpClient = smtp
	got, err = workflow.sendSubscriptionEmail(context.Background(), got)
	if err != nil || smtp.email != "ada@example.com" || !strings.Contains(smtp.body, "N3") {
		t.Fatalf("email = %#v err %v", smtp, err)
	}

	workflow.smtpClient = &fakeSMTP{err: errors.New("smtp down")}
	if _, err = workflow.sendSubscriptionEmail(context.Background(), got); err != nil {
		t.Fatalf("email failure should not fail the paid subscription: %v", err)
	}

	_, err = workflow.paymentFailed(context.Background(), chargedExecution())
	requireRejected(t, err)
	failedPayment := workflow.paymentRepository.(*fakePaymentRepository).updated
	failedSubscription := workflow.subscriptionRepository.(*fakeSubscriptionRepository).updated
	if failedPayment.Status != paymentStatusFailed || failedSubscription.Status != subscriptionStatusFailed {
		t.Fatalf("failed payment %#v subscription %#v", failedPayment, failedSubscription)
	}
}

func readyWorkflow() *paymentWorkflow {
	return &paymentWorkflow{
		workflow:               &fakeWorkflowExecution{},
		paymentRepository:      &fakePaymentRepository{referenceErr: sql.ErrNoRows},
		subscriptionRepository: &fakeSubscriptionRepository{paymentErr: sql.ErrNoRows},
		planRepository:         &fakePlanRepository{plan: monthPlan()},
		userRepository:         &fakeUserRepository{user: &models.User{ID: 4, OnboardingID: 6}},
		onboardingRepository:   &fakeOnboardingRepository{onboarding: &models.Onboarding{ID: 6, Name: "Ada", Email: "ada@example.com"}},
		midtransClient:         &fakeMidtrans{},
		paymentTopic:           &fakePaymentTopic{},
		smtpClient:             &fakeSMTP{},
	}
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
