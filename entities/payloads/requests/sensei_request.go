package requests

import "time"

type (
	Schedule struct {
		Day       string `json:"day"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}

	AssignSenseiClassRequest struct {
		SenseiID    uint64     `json:"sensei_id"`
		Name        string     `json:"name"`
		ReferenceID string     `json:"reference_id"`
		StudentIDs  []uint64   `json:"student_ids"`
		StartDate   time.Time  `json:"start_date"`
		Schedules   []Schedule `json:"schedules"`
	}

	GetSenseiClassScheduleRequest struct {
		Page     int `json:"page"`
		PageSize int `json:"page_size"`
	}

	GetSenseiClassScheduleDetailRequest struct {
		SenseiID uint64 `json:"sensei_id"`
		ClassID  uint64 `json:"class_id"`
	}

	ListOfStudentRecordsRequest struct {
		SenseiID uint64 `json:"sensei_id"`
		ClassID  uint64 `json:"class_id"`
	}

	ConfirmClassPaymentRequest struct {
		TransactionID string `json:"transaction_id"`
	}
)
