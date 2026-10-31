package requests

type (
	AddSubscriptionPlanRequest struct {
		Name          string          `json:"name"`
		Description   string          `json:"description"`
		Price         int             `json:"price"`
		Currency      string          `json:"currency"`
		Interval      string          `json:"interval"`
		IntervalCount int             `json:"interval_count"`
		Features      map[string]bool `json:"features"`
	}

	FeatureFlagEditSubscriptionPlanRequest struct {
		FeatureFlag string `json:"feature_flag"`
		TurnOn      bool   `json:"turn_on"`
	}
)
