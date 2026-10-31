package responses

import (
	"encoding/json"
	"time"
)

type (
	AddSubscriptionPlanResponse struct {
		ID            uint64          `json:"id"`
		Name          string          `json:"name"`
		Description   string          `json:"description"`
		Price         int             `json:"price"`
		Currency      string          `json:"currency"`
		Interval      string          `json:"interval"`
		IntervalCount int             `json:"interval_count"`
		Features      map[string]bool `json:"features"`
		CreatedAt     time.Time       `json:"created_at"`
	}

	GetSubscriptionPlanDetailResponse struct {
		ID            uint64          `json:"id"`
		Name          string          `json:"name"`
		Description   string          `json:"description"`
		Price         int             `json:"price"`
		Currency      string          `json:"currency"`
		Interval      string          `json:"interval"`
		IntervalCount int             `json:"interval_count"`
		Features      json.RawMessage `json:"features"`
		IsActive      bool            `json:"is_active"`
		CreatedAt     time.Time       `json:"created_at"`
		UpdatedAt     time.Time       `json:"updated_at"`
	}

	ListSubscriptionPlansResponse struct {
		SubscriptionPlans []*GetSubscriptionPlanDetailResponse `json:"subscription_plans"`
	}
)
