package responses

import (
	"encoding/json"
	"time"
)

type (
	GenerateExerciseResponse struct {
		ID        uint64          `json:"id"`
		Status    string          `json:"status"`
		JLPTLevel string          `json:"jlpt_level"`
		Exercise  json.RawMessage `json:"exercise,omitempty"`
		CreatedAt time.Time       `json:"created_at"`
		UpdatedAt time.Time       `json:"updated_at"`
	}

	SubmitExerciseResponse struct {
		ID        uint64          `json:"id"`
		Status    string          `json:"status"`
		Answer    json.RawMessage `json:"answer,omitempty"`
		CreatedAt time.Time       `json:"created_at"`
		UpdatedAt time.Time       `json:"updated_at"`
	}
)
