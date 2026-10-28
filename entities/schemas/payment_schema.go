package schemas

import (
	"encoding/json"
	"time"
)

type PaymentSchema struct {
	ID        string          `json:"id"`
	Amount    float64         `json:"amount"`
	Currency  string          `json:"currency"`
	Status    string          `json:"status"`
	PlanID    string          `json:"plan_id"`
	PlanName  string          `json:"plan_name"`
	UserID    string          `json:"user_id"`
	UserName  string          `json:"user_name"`
	UserEmail string          `json:"user_email"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
