package models

import "time"

type ClassRoom struct {
	ID            uint64     `db:"id" json:"id"`
	ClassRoomName string     `db:"class_room_name" json:"class_room_name"`
	RoomCode      string     `db:"room_code" json:"room_code"`
	CreatedBy     uint64     `db:"created_by" json:"created_by"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
