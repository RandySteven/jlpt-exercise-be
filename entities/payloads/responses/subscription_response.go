package responses

type (
	CreateSubscriptionResponse struct {
		SubscriptionID string `json:"subscription_id"`
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
)
