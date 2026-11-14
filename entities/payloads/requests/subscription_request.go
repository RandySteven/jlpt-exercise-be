package requests

import "strings"

type (
	CreateSubscriptionRequest struct {
		IdempotencyKey string `json:"idempotency_key"`
		PlanID         uint64 `json:"plan_id"`
	}

	ChargeSubscriptionRequest struct {
		IdempotencyKey string `json:"idempotency_key"`
		PlanID         uint64 `json:"plan_id"`
	}

	CaptureSubscriptionRequest struct {
		IdempotencyKey string `json:"idempotency_key"`
		PlanID         uint64 `json:"plan_id"`
	}

	PaymentConfirmationRequest struct {
		IdempotencyKey    string `json:"idempotency_key"`
		PlanID            uint64 `json:"plan_id"`
		Status            string `json:"status"`
		TransactionStatus string `json:"transaction_status"`
	}

	CancelSubscriptionRequest struct {
		SubscriptionID uint64 `json:"subscription_id"`
		Reason         string `json:"reason"`
	}

	GetPaymentHistoryRequest struct {
		SubscriptionID uint64 `json:"subscription_id"`
		Page           int    `json:"page"`
		Limit          int    `json:"limit"`
	}

	RefundRequest struct {
		PaymentReferenceID string `json:"payment_reference_id"`
	}
)

func (r *PaymentConfirmationRequest) ResolvedStatus() string {
	if r == nil {
		return ""
	}
	if status := strings.TrimSpace(r.Status); status != "" {
		return status
	}
	return strings.TrimSpace(r.TransactionStatus)
}
