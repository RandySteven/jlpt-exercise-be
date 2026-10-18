package requests

import "encoding/json"

type (
	GenerateExerciseRequest struct {
		IdempotencyKey string `json:"idempotency_key"`
		JLPTNLevel     int    `json:"jlpt_n_level"`
	}

	SubmitExerciseRequest struct {
		ID     uint64          `json:"id"`
		Answer json.RawMessage `json:"answer"`
	}
)
