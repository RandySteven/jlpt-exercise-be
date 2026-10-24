package responses

import "time"

type (
	CreateSubscriptionResponse struct {
		SubscriptionID uint64    `json:"subscription_id"`
		RedirectURL    string    `json:"redirect_url,omitempty"`
		Status         string    `json:"status"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		StartDate      time.Time `json:"start_date"`
		EndDate        time.Time `json:"end_date"`
		PlanID         uint64    `json:"plan_id"`
		PlanName       string    `json:"plan_name"`
	}

	ChargeSubscriptionResponse struct {
		SubscriptionID string `json:"subscription_id"`
	}

	CaptureSubscriptionResponse struct {
		SubscriptionID string `json:"subscription_id"`
	}

	PaymentConfirmationResponse struct {
		RedirectURL string `json:"redirect_url"`
	}

	CancelSubscriptionResponse struct {
		SubscriptionID string `json:"subscription_id"`
	}

	GetPaymentHistoryResponse struct {
		PaymentHistory []PaymentHistory `json:"payment_history"`
		Total          int              `json:"total"`
		Page           int              `json:"page"`
		Limit          int              `json:"limit"`
	}

	PaymentHistory struct {
		PaymentID string    `json:"payment_id"`
		Amount    int       `json:"amount"`
		Currency  string    `json:"currency"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"created_at"`
	}
)
