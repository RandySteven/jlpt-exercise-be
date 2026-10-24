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
)

type fakePaymentUsecase struct {
	createCtx context.Context
	createReq *requests.CreateSubscriptionRequest
	createRes *responses.CreateSubscriptionResponse
	createErr *apperror.CustomError

	confirmReq *requests.PaymentConfirmationRequest
	confirmRes *responses.PaymentConfirmationResponse
	confirmErr *apperror.CustomError

	cancelReq *requests.CancelSubscriptionRequest
	cancelRes *responses.CancelSubscriptionResponse
	cancelErr *apperror.CustomError
}

func (f *fakePaymentUsecase) CreateSubscription(ctx context.Context, request *requests.CreateSubscriptionRequest) (*responses.CreateSubscriptionResponse, *apperror.CustomError) {
	f.createCtx = ctx
	f.createReq = request
	return f.createRes, f.createErr
}

func (f *fakePaymentUsecase) ConfirmPayment(_ context.Context, request *requests.PaymentConfirmationRequest) (*responses.PaymentConfirmationResponse, *apperror.CustomError) {
	f.confirmReq = request
	return f.confirmRes, f.confirmErr
}

func (f *fakePaymentUsecase) CancelSubscription(ctx context.Context, request *requests.CancelSubscriptionRequest) (*responses.CancelSubscriptionResponse, *apperror.CustomError) {
	f.cancelReq = request
	return f.cancelRes, f.cancelErr
}

func (f *fakePaymentUsecase) GetPaymentHistory(context.Context, *requests.GetPaymentHistoryRequest) (*responses.GetPaymentHistoryResponse, *apperror.CustomError) {
	return nil, nil
}

func TestCreateSubscription(t *testing.T) {
	t.Run("rejects invalid json", func(t *testing.T) {
		handler := NewPaymentHandler(&fakePaymentUsecase{})
		req := httptest.NewRequest(http.MethodPost, "/payment/subscription", strings.NewReader(`{`))
		rec := httptest.NewRecorder()

		handler.CreateSubscription(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d body %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("returns the subscription from the usecase", func(t *testing.T) {
		fake := &fakePaymentUsecase{createRes: &responses.CreateSubscriptionResponse{SubscriptionID: 51}}
		handler := NewPaymentHandler(fake)
		req := httptest.NewRequest(http.MethodPost, "/payment/subscription", strings.NewReader(`{"idempotency_key":"key-1","plan_id":3}`))
		req = req.WithContext(context.WithValue(req.Context(), enums.UserID, uint64(7)))
		rec := httptest.NewRecorder()

		handler.CreateSubscription(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body %s", rec.Code, rec.Body.String())
		}
		if fake.createReq == nil || fake.createReq.IdempotencyKey != "key-1" || fake.createReq.PlanID != 3 {
			t.Fatalf("request = %#v", fake.createReq)
		}
		if got, _ := fake.createCtx.Value(enums.UserID).(uint64); got != 7 {
			t.Fatalf("user id = %d", got)
		}
		if !strings.Contains(rec.Body.String(), `"subscription_id":51`) {
			t.Fatalf("body = %s", rec.Body.String())
		}
	})

	t.Run("maps usecase errors", func(t *testing.T) {
		fake := &fakePaymentUsecase{createErr: apperror.NewCustomError(apperror.ErrUnauthorized, "user is missing", nil)}
		handler := NewPaymentHandler(fake)
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
		fake := &fakePaymentUsecase{confirmRes: &responses.PaymentConfirmationResponse{RedirectURL: "https://pay.example/snap"}}
		handler := NewPaymentHandler(fake)
		req := httptest.NewRequest(http.MethodPost, "/payment/confirmation", strings.NewReader(`{"idempotency_key":"key-1","plan_id":3}`))
		rec := httptest.NewRecorder()

		handler.ConfirmPayment(rec, req)

		if rec.Code != http.StatusOK || fake.confirmReq == nil || fake.confirmReq.PlanID != 3 {
			t.Fatalf("status = %d request %#v body %s", rec.Code, fake.confirmReq, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "https://pay.example/snap") {
			t.Fatalf("body = %s", rec.Body.String())
		}
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		handler := NewPaymentHandler(&fakePaymentUsecase{})
		req := httptest.NewRequest(http.MethodPost, "/payment/confirmation", strings.NewReader(`{`))
		rec := httptest.NewRecorder()

		handler.ConfirmPayment(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d body %s", rec.Code, rec.Body.String())
		}
	})
}
