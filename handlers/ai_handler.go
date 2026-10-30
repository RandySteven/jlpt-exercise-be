package handlers

import "net/http"

type (
	IAIHandler interface {
		Chat(w http.ResponseWriter, r *http.Request)
		AvailableModels(w http.ResponseWriter, r *http.Request)
		GetChat(w http.ResponseWriter, r *http.Request)
	}

	AIHandler struct {
	}
)
