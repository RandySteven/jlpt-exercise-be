package models

import (
	"encoding/json"
	"time"
)

type Exercise struct {
	ID          uint64          `db:"id" json:"id"`
	ReferenceID string          `db:"reference_id" json:"reference_id"`
	NLevel      int             `db:"n_level" json:"n_level"`
	Exercise    json.RawMessage `db:"exercise" json:"exercise"`
	Score       int             `db:"score" json:"score"`
	Status      string          `db:"status" json:"status"`
	UserID      uint64          `db:"user_id" json:"user_id"`
	Note        string          `db:"note" json:"note"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
}
