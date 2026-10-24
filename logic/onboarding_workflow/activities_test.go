package onboarding_workflow

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"go.temporal.io/sdk/temporal"
)

func TestPersistRequest(t *testing.T) {
	workflow := &onboardingWorkflow{onboardingRepository: &fakeOnboardingRepository{}}
	_, err := workflow.persistRequest(context.Background(), &ExecutionData{})
	requireInputError(t, err, "register request is missing")

	_, err = workflow.persistRequest(context.Background(), &ExecutionData{Request: &requests.RegisterRequest{
		Password: strings.Repeat("a", 73),
	}})
	requireInputError(t, err, "failed to map onboarding request")

	saveErr := errors.New("insert failed")
	workflow.onboardingRepository = &fakeOnboardingRepository{saveErr: saveErr}
	_, err = workflow.persistRequest(context.Background(), &ExecutionData{Request: registerRequest()})
	if !errors.Is(err, saveErr) {
		t.Fatalf("got %v", err)
	}

	repo := &fakeOnboardingRepository{}
	workflow.onboardingRepository = repo
	got, err := workflow.persistRequest(context.Background(), &ExecutionData{Request: registerRequest()})
	if err != nil {
		t.Fatal(err)
	}
	if got.CurrActivity != registerUserActivity || got.Onboarding == nil || got.Onboarding.ID != 31 || got.Onboarding.Email != "ada@example.com" || got.Onboarding.Status != "PROCESSING" {
		t.Fatalf("persisted = %#v", got.Onboarding)
	}
	if repo.saved == nil || repo.saved.Name != "Ada Lovelace" {
		t.Fatalf("saved = %#v", repo.saved)
	}
}

func TestRegisterUser(t *testing.T) {
	workflow := &onboardingWorkflow{
		userRepository:       &fakeUserRepository{},
		onboardingRepository: &fakeOnboardingRepository{},
	}
	_, err := workflow.registerUser(context.Background(), &ExecutionData{})
	requireInputError(t, err, "register request is missing")
	_, err = workflow.registerUser(context.Background(), &ExecutionData{Request: registerRequest()})
	requireInputError(t, err, "onboarding record is missing")

	saveErr := errors.New("user insert failed")
	workflow.userRepository = &fakeUserRepository{saveErr: saveErr}
	_, err = workflow.registerUser(context.Background(), registerExecution())
	if !errors.Is(err, saveErr) {
		t.Fatalf("got %v", err)
	}

	updateErr := errors.New("onboarding update failed")
	workflow.userRepository = &fakeUserRepository{}
	workflow.onboardingRepository = &fakeOnboardingRepository{updateErr: updateErr}
	_, err = workflow.registerUser(context.Background(), registerExecution())
	if !errors.Is(err, updateErr) {
		t.Fatalf("got %v", err)
	}

	users := &fakeUserRepository{}
	onboarding := &fakeOnboardingRepository{}
	workflow.userRepository = users
	workflow.onboardingRepository = onboarding
	got, err := workflow.registerUser(context.Background(), registerExecution())
	if err != nil {
		t.Fatal(err)
	}
	if users.saved == nil || users.saved.ID != 11 || users.saved.OnboardingID != 31 || users.saved.UserName != "ada" || users.saved.Status != "PROCESSING" {
		t.Fatalf("saved user = %#v", users.saved)
	}
	if onboarding.updated == nil || onboarding.updated.Status != "ACTIVE" || got.Onboarding.Status != "ACTIVE" {
		t.Fatalf("onboarding = %#v", got.Onboarding)
	}
	if got.CurrActivity != sendActivationEmailActivity || got.ActivationToken == "" || got.Response == nil || got.Response.UserID != 11 || got.Response.UserEmail != "ada@example.com" || got.Response.ActivationToken != got.ActivationToken {
		t.Fatalf("execution = %#v", got)
	}
	claims, err := decodeActivationToken(got.ActivationToken)
	if err != nil || claims.UserID != 11 || claims.OnboardingID != 31 || claims.WorkflowID != "wf" || claims.RunID != "run" {
		t.Fatalf("claims = %#v err = %v", claims, err)
	}
}

func TestUpdateOnboardingStatus(t *testing.T) {
	workflow := &onboardingWorkflow{
		onboardingRepository: &fakeOnboardingRepository{},
		userRepository:       &fakeUserRepository{},
	}
	_, err := workflow.updateOnboardingStatus(context.Background(), &ExecutionData{})
	requireInputError(t, err, "onboarding record is missing")

	updateErr := errors.New("update failed")
	workflow.onboardingRepository = &fakeOnboardingRepository{updateErr: updateErr}
	_, err = workflow.updateOnboardingStatus(context.Background(), &ExecutionData{Onboarding: &models.Onboarding{ID: 31}})
	if !errors.Is(err, updateErr) {
		t.Fatalf("got %v", err)
	}

	workflow.onboardingRepository = &fakeOnboardingRepository{}
	userErr := errors.New("user update failed")
	workflow.userRepository = &fakeUserRepository{updateErr: userErr}
	_, err = workflow.updateOnboardingStatus(context.Background(), &ExecutionData{
		Onboarding: &models.Onboarding{ID: 31, Name: "Ada"},
		User:       &models.User{ID: 11, UserName: "ada"},
	})
	if !errors.Is(err, userErr) {
		t.Fatalf("got %v", err)
	}

	users := &fakeUserRepository{}
	workflow.userRepository = users
	got, err := workflow.updateOnboardingStatus(context.Background(), &ExecutionData{Onboarding: &models.Onboarding{ID: 31, Name: "Ada", Email: "ada@example.com"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Onboarding.Status != "ONBOARDED" || got.Response == nil || got.Response.OnboardingID != 31 || users.updated != nil {
		t.Fatalf("onboarding %#v response %#v user %#v", got.Onboarding, got.Response, users.updated)
	}

	got, err = workflow.updateOnboardingStatus(context.Background(), &ExecutionData{
		Onboarding:      &models.Onboarding{ID: 31, Name: "Ada", Email: "ada@example.com"},
		User:            &models.User{ID: 11, UserName: "ada"},
		ActivationToken: "token",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.User.Status != "ACTIVE" || got.Response.UserID != 11 || got.Response.UserName != "ada" || got.Response.ActivationToken != "token" {
		t.Fatalf("user %#v response %#v", got.User, got.Response)
	}
}

func TestSendActivationEmail(t *testing.T) {
	workflow := &onboardingWorkflow{}
	_, err := workflow.sendActivationEmail(context.Background(), &ExecutionData{})
	requireInputError(t, err, "onboarding email is missing")

	_, err = workflow.sendActivationEmail(context.Background(), &ExecutionData{Onboarding: &models.Onboarding{Email: "ada@example.com"}})
	requireInputError(t, err, "activation token is missing")

	_, err = workflow.sendActivationEmail(context.Background(), &ExecutionData{
		Onboarding:      &models.Onboarding{Email: "ada@example.com"},
		ActivationToken: "token",
	})
	requireInputError(t, err, "smtp client is missing")

	sendErr := errors.New("smtp down")
	workflow.smtpClient = &fakeSMTP{err: sendErr}
	_, err = workflow.sendActivationEmail(context.Background(), &ExecutionData{
		Onboarding:      &models.Onboarding{Email: "ada@example.com"},
		ActivationToken: "token",
	})
	if !errors.Is(err, sendErr) {
		t.Fatalf("got %v", err)
	}

	smtp := &fakeSMTP{}
	workflow.smtpClient = smtp
	got, err := workflow.sendActivationEmail(context.Background(), &ExecutionData{
		Onboarding:      &models.Onboarding{Email: "ada@example.com"},
		ActivationToken: "token",
	})
	if err != nil {
		t.Fatal(err)
	}
	if smtp.email != "ada@example.com" || smtp.subject != activationEmailSubject || smtp.template != activationEmailTemplate {
		t.Fatalf("smtp = %#v", smtp)
	}
	data, ok := smtp.data.(map[string]string)
	if !ok || data["onboarding_email"] != "ada@example.com" || data["activation_token"] != "token" {
		t.Fatalf("template data = %#v", smtp.data)
	}
	if got.CurrActivity != updateOnboardingStatusActivity {
		t.Fatalf("next activity = %q", got.CurrActivity)
	}
}

func registerRequest() *requests.RegisterRequest {
	return &requests.RegisterRequest{
		FirstName:   "Ada",
		LastName:    "Lovelace",
		Email:       "ada@example.com",
		Username:    "ada",
		Password:    "secret",
		PhoneNumber: "0800",
		Address:     "London",
	}
}

func registerExecution() *ExecutionData {
	return &ExecutionData{
		Request:    registerRequest(),
		Onboarding: &models.Onboarding{ID: 31, Name: "Ada Lovelace", Email: "ada@example.com", Status: "PROCESSING", RegisterAs: enums.RegisterAsUser},
		WorkflowID: "wf",
		RunID:      "run",
	}
}

func requireInputError(t *testing.T, err error, message string) {
	t.Helper()
	var appErr *temporal.ApplicationError
	if !errors.As(err, &appErr) || appErr.Type() != onboardingInputErrorType || !appErr.NonRetryable() || appErr.Message() != message {
		t.Fatalf("got %v, want non-retryable %s %q", err, onboardingInputErrorType, message)
	}
}

func TestPendingActivationResponse(t *testing.T) {
	data := &ExecutionData{ActivationToken: "token"}
	data.setPendingActivationResponse()
	if data.Response == nil || data.Response.ActivationToken != "token" || data.Response.UserID != 0 {
		t.Fatalf("token only = %#v", data.Response)
	}

	data.Onboarding = &models.Onboarding{ID: 3, Name: "Ada Lovelace", Email: "ada@example.com"}
	data.User = &models.User{ID: 8}
	data.setPendingActivationResponse()
	if data.Response.UserName != "Ada Lovelace" || data.Response.UserEmail != "ada@example.com" || data.Response.UserID != 8 {
		t.Fatalf("onboarding name = %#v", data.Response)
	}

	data.User.UserName = "ada"
	data.setPendingActivationResponse()
	if data.Response.UserName != "ada" {
		t.Fatalf("username = %#v", data.Response)
	}

	raw := json.RawMessage(`{"city":"London"}`)
	mapped := mappingUserInfo(raw)
	if string(mapped.AdditionalInfo) != string(raw) {
		t.Fatalf("additional info = %s", mapped.AdditionalInfo)
	}
}
