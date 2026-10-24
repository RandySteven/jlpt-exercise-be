package models

import (
	"encoding/json"
	"time"
)

type SubscriptionPlan struct {
	ID            uint64          `db:"id" json:"id"`
	Name          string          `db:"name" json:"name"`
	Price         int             `db:"price" json:"price"`
	Currency      string          `db:"currency" json:"currency"`
	Interval      string          `db:"interval" json:"interval"`
	IntervalCount int             `db:"interval_count" json:"interval_count"`
	FeatureFlags  json.RawMessage `db:"feature_flags" json:"feature_flags"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at" json:"updated_at"`
	DeletedAt     *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
}
