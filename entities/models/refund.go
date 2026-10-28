package models

import (
	"encoding/json"
	"time"
)

type Refund struct {
	ID                  uint64          `db:"id" json:"id"`
	ReferenceID         string          `db:"reference_id" json:"reference_id"`
	OriginalReferenceID string          `db:"original_reference_id" json:"original_reference_id"`
	PaymentID           uint64          `db:"payment_id" json:"payment_id"`
	UserID              uint64          `db:"user_id" json:"user_id"`
	Amount              int             `db:"amount" json:"amount"`
	Status              string          `db:"status" json:"status"`
	Reason              string          `db:"reason" json:"reason"`
	Metadata            json.RawMessage `db:"metadata" json:"metadata"`
	CreatedAt           time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time       `db:"updated_at" json:"updated_at"`
	DeletedAt           *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
}
