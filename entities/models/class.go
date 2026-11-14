package models

import "time"

type ClassSchedule struct {
	Day       string `db:"day" json:"day"`
	StartTime string `db:"start_time" json:"start_time"`
	EndTime   string `db:"end_time" json:"end_time"`
}

type Class struct {
	ID          uint64          `db:"id" json:"id"`
	SenseiID    uint64          `db:"sensei_id" json:"sensei_id"`
	Name        string          `db:"name" json:"name"`
	ReferenceID string          `db:"reference_id" json:"reference_id"`
	Schedules   []ClassSchedule `db:"schedules" json:"schedules"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
}
