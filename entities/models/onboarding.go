package models

import (
	"encoding/json"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/enums"
)

type Onboarding struct {
	ID                uint64           `db:"id" json:"id"`
	Name              string           `db:"name" json:"name"`
	Email             string           `db:"email" json:"email"`
	Password          string           `db:"password" json:"password"`
	PhoneNumber       string           `db:"phone_number" json:"phone_number"`
	Address           string           `db:"address" json:"address"`
	RegisterAs        enums.RegisterAs `db:"register_as" json:"register_as"`
	Status            string           `db:"status" json:"status"`
	ProfilePictureURL *string          `db:"profile_picture_url" json:"profile_picture_url,omitempty"`
	AdditionalInfo    json.RawMessage  `db:"additional_info" json:"additional_info"`
	CreatedAt         time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time        `db:"updated_at" json:"updated_at"`
	DeletedAt         *time.Time       `db:"deleted_at" json:"deleted_at,omitempty"`
}
