package requests

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
		IdempotencyKey string `json:"idempotency_key"`
		PlanID         uint64 `json:"plan_id"`
		Status         string `json:"status"`
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
