package handlers

import (
	"net/http"

	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/usecases"
	"github.com/RandySteven/jlpt-exercise-be/utils"
)

type (
	IPaymentHandler interface {
		CreateSubscription(w http.ResponseWriter, r *http.Request)
		ConfirmPayment(w http.ResponseWriter, r *http.Request)
		CancelSubscription(w http.ResponseWriter, r *http.Request)
		GetSubscription(w http.ResponseWriter, r *http.Request)
		GetPaymentHistory(w http.ResponseWriter, r *http.Request)
	}

	PaymentHandler struct {
		paymentUsecase usecases.PaymentUsecase
	}
)

// CancelSubscription implements [IPaymentHandler].
func (p *PaymentHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	var req requests.CancelSubscriptionRequest
	dataKey := "cancel_subscription_response"
	if err := utils.BindJSON(r, &req); err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, "Invalid request body", nil, nil, err)
		return
	}

	response, customErr := p.paymentUsecase.CancelSubscription(r.Context(), &req)
	if customErr != nil {
		utils.ResponseHandler(w, customErr.ErrCode(), "Failed to cancel subscription", nil, nil, customErr)
		return
	}

	utils.ResponseHandler(w, http.StatusOK, "Subscription cancelled successfully", &dataKey, response, nil)
}

// GetPaymentHistory implements [IPaymentHandler].
func (p *PaymentHandler) GetPaymentHistory(w http.ResponseWriter, r *http.Request) {
	dataKey := "payment_history_response"
	var req requests.GetPaymentHistoryRequest
	if err := utils.BindJSON(r, &req); err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, "Invalid request body", nil, nil, err)
		return
	}

	response, customErr := p.paymentUsecase.GetPaymentHistory(r.Context(), &req)
	if customErr != nil {
		utils.ResponseHandler(w, customErr.ErrCode(), "Failed to get payment history", nil, nil, customErr)
		return
	}

	utils.ResponseHandler(w, http.StatusOK, "Payment history retrieved successfully", &dataKey, response, nil)
}

// GetSubscription implements [IPaymentHandler].
func (p *PaymentHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	dataKey := "subscription_detail_response"
	response, customErr := p.paymentUsecase.GetUserSubscriptionDetail(r.Context())
	if customErr != nil {
		utils.ResponseHandler(w, customErr.ErrCode(), "Failed to get subscription detail", nil, nil, customErr)
		return
	}

	utils.ResponseHandler(w, http.StatusOK, "Subscription detail retrieved successfully", &dataKey, response, nil)
}

func (p *PaymentHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req requests.CreateSubscriptionRequest
	dataKey := "create_subscription_response"
	if err := utils.BindJSON(r, &req); err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, "Invalid request body", nil, nil, err)
		return
	}

	response, customErr := p.paymentUsecase.CreateSubscription(r.Context(), &req)
	if customErr != nil {
		utils.ResponseHandler(w, customErr.ErrCode(), "Failed to create subscription", nil, nil, customErr)
		return
	}

	utils.ResponseHandler(w, http.StatusOK, "Subscription created successfully", &dataKey, response, nil)
}

func (p *PaymentHandler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	var req requests.PaymentConfirmationRequest
	dataKey := "payment_confirmation_response"
	if err := utils.BindJSON(r, &req); err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, "Invalid request body", nil, nil, err)
		return
	}

	response, customErr := p.paymentUsecase.ConfirmPayment(r.Context(), &req)
	if customErr != nil {
		utils.ResponseHandler(w, customErr.ErrCode(), "Failed to confirm payment", nil, nil, customErr)
		return
	}

	utils.ResponseHandler(w, http.StatusOK, "Payment confirmed successfully", &dataKey, response, nil)
}

var _ IPaymentHandler = &PaymentHandler{}

func NewPaymentHandler(paymentUsecase usecases.PaymentUsecase) IPaymentHandler {
	return &PaymentHandler{
		paymentUsecase: paymentUsecase,
	}
}
