package usecases

import (
	"context"

	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

type (
	AIUsecase interface {
		Chat(ctx context.Context, request *requests.AIChatRequest) (*responses.AIChatResponse, error)
		ListModels(ctx context.Context) (*responses.ListModelsResponse, error)
		ListChatRooms(ctx context.Context) (*responses.ListChatRoomsResponse, error)
		GetChat(ctx context.Context, request *requests.GetChatRequest) error
	}
)
