package responses

import "time"

type (
	ListClassResponse struct {
	}

	StudentResponse struct {
		StudentID uint64 `json:"student_id"`
		Name      string `json:"name"`
		Status    string `json:"status"`
	}

	AssignSenseiClassResponse struct {
		ClassID      uint64             `json:"class_id"`
		Status       string             `json:"status"`
		SuccessCount uint64             `json:"success_count"`
		CreatedAt    time.Time          `json:"created_at"`
		Students     []*StudentResponse `json:"students"`
	}

	GetSenseiClassScheduleResponse struct {
		Classes  []ListClassResponse `json:"classes"`
		Total    int                 `json:"total"`
		Page     int                 `json:"page"`
		PageSize int                 `json:"page_size"`
	}

	GetSenseiClassScheduleDetailResponse struct {
	}

	ConfirmClassPaymentResponse struct {
		TransactionID string `json:"transaction_id"`
	}
)
