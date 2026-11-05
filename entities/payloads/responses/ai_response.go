package responses

type (
	AIChatResponse struct {
		Message      string `json:"message"`
		ResponseTime int64  `json:"response_time"`
	}

	ListModelsResponse struct {
		Models []Model `json:"models"`
	}

	Model struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	ListChatRoomsResponse struct {
		ChatRooms []ChatRoom `json:"chat_rooms"`
	}

	ChatRoom struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	ListChatMessagesResponse struct {
		Messages []ChatMessage `json:"messages"`
	}

	ChatMessage struct {
		ID      string `json:"id"`
		Role    string `json:"role"`
		Content string `json:"content"`
	}
)
