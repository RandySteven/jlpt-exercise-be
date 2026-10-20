package requests

type (
	CreateSubscriptionRequest struct {
		IdempotencyKey string `json:"idempotency_key"`
		PlanID         uint64 `json:"plan_id"`
	}
)
