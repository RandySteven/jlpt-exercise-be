package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	usecasemocks "github.com/RandySteven/jlpt-exercise-be/usecases/mocks"
	"github.com/stretchr/testify/mock"
)

func TestCreateSubscription(t *testing.T) {
	t.Run("rejects invalid json", func(t *testing.T) {
		handler := NewPaymentHandler(usecasemocks.NewMockPaymentUsecase(t))
		req := httptest.NewRequest(http.MethodPost, "/payment/subscription", strings.NewReader(`{`))
		rec := httptest.NewRecorder()

		handler.CreateSubscription(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d body %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("returns the subscription from the usecase", func(t *testing.T) {
		usecase := usecasemocks.NewMockPaymentUsecase(t)
		var gotCtx context.Context
		var gotReq *requests.CreateSubscriptionRequest
		usecase.EXPECT().CreateSubscription(mock.Anything, mock.Anything).Run(func(ctx context.Context, request *requests.CreateSubscriptionRequest) {
			gotCtx = ctx
			gotReq = request
		}).Return(&responses.CreateSubscriptionResponse{SubscriptionID: 51}, nil)
		handler := NewPaymentHandler(usecase)
		req := httptest.NewRequest(http.MethodPost, "/payment/subscription", strings.NewReader(`{"idempotency_key":"key-1","plan_id":3}`))
		req = req.WithContext(context.WithValue(req.Context(), enums.UserID, uint64(7)))
		rec := httptest.NewRecorder()

		handler.CreateSubscription(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body %s", rec.Code, rec.Body.String())
		}
		if gotReq == nil || gotReq.IdempotencyKey != "key-1" || gotReq.PlanID != 3 {
			t.Fatalf("request = %#v", gotReq)
		}
		if got, _ := gotCtx.Value(enums.UserID).(uint64); got != 7 {
			t.Fatalf("user id = %d", got)
		}
		if !strings.Contains(rec.Body.String(), `"subscription_id":51`) {
			t.Fatalf("body = %s", rec.Body.String())
		}
	})

	t.Run("maps usecase errors", func(t *testing.T) {
		usecase := usecasemocks.NewMockPaymentUsecase(t)
		usecase.EXPECT().CreateSubscription(mock.Anything, mock.Anything).Return(nil, apperror.NewCustomError(apperror.ErrUnauthorized, "user is missing", nil))
		handler := NewPaymentHandler(usecase)
		req := httptest.NewRequest(http.MethodPost, "/payment/subscription", strings.NewReader(`{"idempotency_key":"key-1","plan_id":3}`))
		rec := httptest.NewRecorder()

		handler.CreateSubscription(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d body %s", rec.Code, rec.Body.String())
		}
	})
}

func TestConfirmPayment(t *testing.T) {
	t.Run("returns the redirect url", func(t *testing.T) {
		usecase := usecasemocks.NewMockPaymentUsecase(t)
		var gotReq *requests.PaymentConfirmationRequest
		usecase.EXPECT().ConfirmPayment(mock.Anything, mock.Anything).Run(func(_ context.Context, request *requests.PaymentConfirmationRequest) {
			gotReq = request
		}).Return(&responses.PaymentConfirmationResponse{RedirectURL: "https://pay.example/snap"}, nil)
		handler := NewPaymentHandler(usecase)
		req := httptest.NewRequest(http.MethodPost, "/payment/confirmation", strings.NewReader(`{"idempotency_key":"key-1","plan_id":3}`))
		rec := httptest.NewRecorder()

		handler.ConfirmPayment(rec, req)

		if rec.Code != http.StatusOK || gotReq == nil || gotReq.PlanID != 3 {
			t.Fatalf("status = %d request %#v body %s", rec.Code, gotReq, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "https://pay.example/snap") {
			t.Fatalf("body = %s", rec.Body.String())
		}
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		handler := NewPaymentHandler(usecasemocks.NewMockPaymentUsecase(t))
		req := httptest.NewRequest(http.MethodPost, "/payment/confirmation", strings.NewReader(`{`))
		rec := httptest.NewRecorder()

		handler.ConfirmPayment(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d body %s", rec.Code, rec.Body.String())
		}
	})
}
