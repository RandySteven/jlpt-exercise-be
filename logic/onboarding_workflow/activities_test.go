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
	smtpmocks "github.com/RandySteven/jlpt-exercise-be/pkg/smtp/mocks"
	repomocks "github.com/RandySteven/jlpt-exercise-be/repositories/mocks"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"
)

func TestPersistRequest(t *testing.T) {
	onboarding := repomocks.NewMockOnboardingRepository(t)
	workflow := &onboardingWorkflow{onboardingRepository: onboarding}
	_, err := workflow.persistRequest(context.Background(), &ExecutionData{})
	requireInputError(t, err, "register request is missing")

	_, err = workflow.persistRequest(context.Background(), &ExecutionData{Request: &requests.RegisterRequest{
		Password: strings.Repeat("a", 73),
	}})
	requireInputError(t, err, "failed to map onboarding request")

	saveErr := errors.New("insert failed")
	onboarding.EXPECT().Save(mock.Anything, mock.Anything).Return(nil, saveErr).Once()
	_, err = workflow.persistRequest(context.Background(), &ExecutionData{Request: registerRequest()})
	if !errors.Is(err, saveErr) {
		t.Fatalf("got %v", err)
	}

	var saved *models.Onboarding
	onboarding.EXPECT().Save(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, entity *models.Onboarding) (*models.Onboarding, error) {
		copied := *entity
		if copied.ID == 0 {
			copied.ID = 31
		}
		saved = &copied
		return saved, nil
	}).Once()
	got, err := workflow.persistRequest(context.Background(), &ExecutionData{Request: registerRequest()})
	if err != nil {
		t.Fatal(err)
	}
	if got.CurrActivity != registerUserActivity || got.Onboarding == nil || got.Onboarding.ID != 31 || got.Onboarding.Email != "ada@example.com" || got.Onboarding.Status != "PROCESSING" {
		t.Fatalf("persisted = %#v", got.Onboarding)
	}
	if saved == nil || saved.Name != "Ada Lovelace" {
		t.Fatalf("saved = %#v", saved)
	}
}

func TestRegisterUser(t *testing.T) {
	users := repomocks.NewMockUserRepository(t)
	onboarding := repomocks.NewMockOnboardingRepository(t)
	workflow := &onboardingWorkflow{userRepository: users, onboardingRepository: onboarding}
	_, err := workflow.registerUser(context.Background(), &ExecutionData{})
	requireInputError(t, err, "register request is missing")
	_, err = workflow.registerUser(context.Background(), &ExecutionData{Request: registerRequest()})
	requireInputError(t, err, "onboarding record is missing")

	saveErr := errors.New("user insert failed")
	users.EXPECT().Save(mock.Anything, mock.Anything).Return(nil, saveErr).Once()
	_, err = workflow.registerUser(context.Background(), registerExecution())
	if !errors.Is(err, saveErr) {
		t.Fatalf("got %v", err)
	}

	updateErr := errors.New("onboarding update failed")
	var saved *models.User
	users.EXPECT().Save(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, entity *models.User) (*models.User, error) {
		copied := *entity
		if copied.ID == 0 {
			copied.ID = 11
		}
		saved = &copied
		return saved, nil
	})
	onboarding.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, updateErr).Once()
	_, err = workflow.registerUser(context.Background(), registerExecution())
	if !errors.Is(err, updateErr) {
		t.Fatalf("got %v", err)
	}

	var updated *models.Onboarding
	onboarding.EXPECT().Update(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, entity *models.Onboarding) (*models.Onboarding, error) {
		copied := *entity
		updated = &copied
		return updated, nil
	}).Once()
	got, err := workflow.registerUser(context.Background(), registerExecution())
	if err != nil {
		t.Fatal(err)
	}
	if saved == nil || saved.ID != 11 || saved.OnboardingID != 31 || saved.UserName != "ada" || saved.Status != "PROCESSING" {
		t.Fatalf("saved user = %#v", saved)
	}
	if updated == nil || updated.Status != "ACTIVE" || got.Onboarding.Status != "ACTIVE" {
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
	onboarding := repomocks.NewMockOnboardingRepository(t)
	users := repomocks.NewMockUserRepository(t)
	workflow := &onboardingWorkflow{onboardingRepository: onboarding, userRepository: users}
	_, err := workflow.updateOnboardingStatus(context.Background(), &ExecutionData{})
	requireInputError(t, err, "onboarding record is missing")

	updateErr := errors.New("update failed")
	onboarding.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, updateErr).Once()
	_, err = workflow.updateOnboardingStatus(context.Background(), &ExecutionData{Onboarding: &models.Onboarding{ID: 31}})
	if !errors.Is(err, updateErr) {
		t.Fatalf("got %v", err)
	}

	onboarding.EXPECT().Update(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, entity *models.Onboarding) (*models.Onboarding, error) {
		copied := *entity
		return &copied, nil
	})
	userErr := errors.New("user update failed")
	users.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, userErr).Once()
	_, err = workflow.updateOnboardingStatus(context.Background(), &ExecutionData{
		Onboarding: &models.Onboarding{ID: 31, Name: "Ada"},
		User:       &models.User{ID: 11, UserName: "ada"},
	})
	if !errors.Is(err, userErr) {
		t.Fatalf("got %v", err)
	}

	var updatedUser *models.User
	users.EXPECT().Update(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, entity *models.User) (*models.User, error) {
		copied := *entity
		updatedUser = &copied
		return updatedUser, nil
	}).Maybe()
	got, err := workflow.updateOnboardingStatus(context.Background(), &ExecutionData{Onboarding: &models.Onboarding{ID: 31, Name: "Ada", Email: "ada@example.com"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Onboarding.Status != "ONBOARDED" || got.Response == nil || got.Response.OnboardingID != 31 || updatedUser != nil {
		t.Fatalf("onboarding %#v response %#v user %#v", got.Onboarding, got.Response, updatedUser)
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
	smtp := smtpmocks.NewMockSmtpClient(t)
	smtp.EXPECT().SendHTMLTemplate(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sendErr).Once()
	workflow.smtpClient = smtp
	_, err = workflow.sendActivationEmail(context.Background(), &ExecutionData{
		Onboarding:      &models.Onboarding{Email: "ada@example.com"},
		ActivationToken: "token",
	})
	if !errors.Is(err, sendErr) {
		t.Fatalf("got %v", err)
	}

	var email, subject, template string
	var data any
	smtp.EXPECT().SendHTMLTemplate(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(gotEmail string, gotSubject string, templateName string, gotData any) {
			email = gotEmail
			subject = gotSubject
			template = templateName
			data = gotData
		}).Return(nil).Once()
	got, err := workflow.sendActivationEmail(context.Background(), &ExecutionData{
		Onboarding:      &models.Onboarding{Email: "ada@example.com"},
		ActivationToken: "token",
	})
	if err != nil {
		t.Fatal(err)
	}
	if email != "ada@example.com" || subject != activationEmailSubject || template != activationEmailTemplate {
		t.Fatalf("smtp email=%s subject=%s template=%s", email, subject, template)
	}
	templateData, ok := data.(map[string]string)
	if !ok || templateData["onboarding_email"] != "ada@example.com" || templateData["activation_token"] != "token" {
		t.Fatalf("template data = %#v", data)
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
