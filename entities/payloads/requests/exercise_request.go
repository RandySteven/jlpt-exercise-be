package requests

type (
	GenerateExerciseRequest struct {
		IdempotencyKey string `json:"idempotency_key"`
		JLPTNLevel     int    `json:"jlpt_n_level"`
	}
)
