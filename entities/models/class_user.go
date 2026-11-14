package models

import "time"

type ClassUser struct {
	ID        uint64     `db:"id" json:"id"`
	ClassID   uint64     `db:"class_id" json:"class_id"`
	UserID    uint64     `db:"user_id" json:"user_id"`
	Status    string     `db:"status" json:"status"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
