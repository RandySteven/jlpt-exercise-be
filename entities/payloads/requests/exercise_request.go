package requests

type (
	GenerateExerciseRequest struct {
		IdempotencyKey string `json:"idempotency_key"`
		JLPTNLevel     int    `json:"jlpt_n_level"`
	}

	SubmitExerciseRequest struct {
		ID     uint64                    `json:"id"`
		Answer map[string]map[int]string `json:"answer"`
	}

	ListExerciseRequest struct {
		Page     int64 `json:"page"`
		PageSize int64 `json:"page_size"`
	}
)
