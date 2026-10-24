package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RandySteven/jlpt-exercise-be/handlers"
	"github.com/gorilla/mux"
)

func TestLoginPreflightAllowsGitHubPagesOrigin(t *testing.T) {
	router := mux.NewRouter()
	InitRouter(NewEndpointRouters(&handlers.Handlers{
		OnboardingHandler: handlers.NewOnboardingHandler(nil),
		DummyHandler:      handlers.NewDummyHandler(nil),
		ExerciseHandler:   handlers.NewExerciseHandler(nil),
		PaymentHandler:    handlers.NewPaymentHandler(nil),
		WebhookHandler:    handlers.NewWebhookHandler(nil),
	}), router)

	req := httptest.NewRequest(http.MethodOptions, "/auth/login", nil)
	req.Header.Set("Origin", "https://randysteven.github.io")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://randysteven.github.io" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("Access-Control-Allow-Methods is empty")
	}
}
