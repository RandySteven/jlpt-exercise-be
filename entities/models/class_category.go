package models

import "time"

type ClassCategory struct {
	ID          uint64     `db:"id" json:"id"`
	Category    string     `db:"category" json:"category"`
	Description string     `db:"description" json:"description"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at"`
}
