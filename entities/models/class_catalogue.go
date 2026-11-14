package models

import (
	"encoding/json"
	"time"
)

type ClassCatalogue struct {
	ID             uint64          `db:"id" json:"id"`
	Name           string          `db:"name" json:"name"`
	CategoryID     uint64          `db:"category_id" json:"category_id"`
	Description    string          `db:"description" json:"description"`
	AdditionalInfo json.RawMessage `db:"additional_info" json:"additional_info"`
	Price          int             `db:"price" json:"price"`
	Currency       string          `db:"currency" json:"currency"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time      `db:"deleted_at" json:"deleted_at"`
}
