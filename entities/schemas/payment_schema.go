package schemas

import (
	"encoding/json"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/enums"
)

type PaymentSchema struct {
	ID              string                `json:"id"`
	Amount          float64               `json:"amount"`
	Currency        string                `json:"currency"`
	Status          string                `json:"status"`
	TransactionID   string                `json:"transaction_id"`
	TransactionType enums.TransactionType `json:"transaction_type"`
	UserID          string                `json:"user_id"`
	UserName        string                `json:"user_name"`
	UserEmail       string                `json:"user_email"`
	Metadata        json.RawMessage       `json:"metadata"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}
