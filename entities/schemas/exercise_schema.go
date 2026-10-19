package schemas

import (
	"encoding/json"
	"time"
)

type (
	ExerciseSchema struct {
		ID        uint64          `json:"id"`
		UserID    uint64          `json:"user_id"`
		Exercise  json.RawMessage `json:"exercise"`
		Answer    json.RawMessage `json:"answer"`
		NLevel    int             `json:"n_level"`
		Note      string          `json:"note"`
		Score     int             `json:"score"`
		Status    string          `json:"status"`
		CreatedAt time.Time       `json:"created_at"`
		UpdatedAt time.Time       `json:"updated_at"`
	}
)
