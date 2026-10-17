package models

import (
	"encoding/json"
	"time"
)

type User struct {
	ID             uint64          `db:"id" json:"id"`
	OnboardingID   uint64          `db:"onboarding_id" json:"onboarding_id"`
	UserName       string          `db:"user_name" json:"user_name"`
	Status         string          `db:"status" json:"status"`
	AdditionalInfo json.RawMessage `db:"additional_info" json:"additional_info"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
}
