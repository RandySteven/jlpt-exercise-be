package requests

type (
	AIChatRequest struct {
		Model   string `json:"model"`
		Message string `json:"message"`
	}

	GetChatRequest struct {
		RoomID string `json:"room_id"`
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
	}
)
