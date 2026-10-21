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
	}
)
