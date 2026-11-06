package responses

type (
	ListClassResponse struct {
	}

	AssignSenseiClassResponse struct {
		ClassID uint64 `json:"class_id"`
	}

	GetSenseiClassScheduleResponse struct {
		Classes  []ListClassResponse `json:"classes"`
		Total    int                 `json:"total"`
		Page     int                 `json:"page"`
		PageSize int                 `json:"page_size"`
	}

	GetSenseiClassScheduleDetailResponse struct {
	}
)
