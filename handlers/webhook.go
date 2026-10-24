package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/usecases"
)

type (
	MidtransCallbackRequest struct {
		TransactionTime        string `json:"transaction_time"`
		TransactionStatus      string `json:"transaction_status"`
		TransactionID          string `json:"transaction_id"`
		StatusMessage          string `json:"status_message"`
		StatusCode             string `json:"status_code"`
		SignatureKey           string `json:"signature_key"`
		PaymentType            string `json:"payment_type"`
		OrderID                string `json:"order_id"`
		MerchantID             string `json:"merchant_id"`
		MaskedCard             string `json:"masked_card"`
		GrossAmount            string `json:"gross_amount"`
		FraudStatus            string `json:"fraud_status"`
		ECI                    string `json:"eci"`
		Currency               string `json:"currency"`
		ChannelResponseMessage string `json:"channel_response_message"`
		ChannelResponseCode    string `json:"channel_response_code"`
		CardType               string `json:"card_type"`
		Bank                   string `json:"bank"`
		ApprovalCode           string `json:"approval_code"`
	}

	IWebhookHandler interface {
		Callback(w http.ResponseWriter, r *http.Request)
	}

	WebhookHandler struct {
		paymentUsecase usecases.PaymentUsecase
	}
)

func (h *WebhookHandler) Callback(w http.ResponseWriter, r *http.Request) {
	var request MidtransCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println(request)
	h.paymentUsecase.ConfirmPayment(context.Background(), &requests.PaymentConfirmationRequest{
		IdempotencyKey: request.OrderID,
		Status:         request.TransactionStatus,
	})
}

func NewWebhookHandler(paymentUsecase usecases.PaymentUsecase) *WebhookHandler {
	return &WebhookHandler{
		paymentUsecase: paymentUsecase,
	}
}
