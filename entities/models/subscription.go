package models

import "time"

type Subscription struct {
	ID          uint64     `db:"id" json:"id"`
	ReferenceID string     `db:"reference_id" json:"reference_id"`
	UserID      uint64     `db:"user_id" json:"user_id"`
	PlanID      uint64     `db:"plan_id" json:"plan_id"`
	StartDate   time.Time  `db:"start_date" json:"start_date"`
	EndDate     time.Time  `db:"end_date" json:"end_date"`
	Status      string     `db:"status" json:"status"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
