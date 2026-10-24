package models

import (
	"encoding/json"
	"time"
)

type Payment struct {
	ID             uint64          `db:"id" json:"id"`
	ReferenceID    string          `db:"reference_id" json:"reference_id"`
	SubscriptionID uint64          `db:"subscription_id" json:"subscription_id"`
	UserID         uint64          `db:"user_id" json:"user_id"`
	Amount         int             `db:"amount" json:"amount"`
	Currency       string          `db:"currency" json:"currency"`
	Status         string          `db:"status" json:"status"`
	Metadata       json.RawMessage `db:"metadata" json:"metadata"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
}
