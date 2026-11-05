package usecases

import (
	"context"
	"strings"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	openai_client "github.com/RandySteven/jlpt-exercise-be/pkg/openai"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/utils"
)

type (
	AIUsecase interface {
		Chat(ctx context.Context, request *requests.AIChatRequest) (*responses.AIChatResponse, error)
		StreamChat(ctx context.Context, request *requests.AIChatRequest) (<-chan openai_client.StreamChunk, error)
		ListModels(ctx context.Context) (*responses.ListModelsResponse, error)
		ListChatRooms(ctx context.Context) (*responses.ListChatRoomsResponse, error)
		GetChat(ctx context.Context, request *requests.GetChatRequest) (*responses.ListChatMessagesResponse, error)
	}

	aiUsecase struct {
		subscriptionRepository     repositories.SubscriptionRepository
		subscriptionPlanRepository repositories.SubscriptionPlanRepository
		openaiClient               openai_client.OpenAIClient
	}
)

const senseiInstructions = `You are a JLPT Japanese tutor. Help only with Japanese study: grammar, vocabulary, kanji, reading, corrections, and study plans. If the user asks for something else, refuse in one short sentence and offer to turn it into Japanese practice.`

// Chat implements [AIUsecase].
func (a *aiUsecase) Chat(ctx context.Context, request *requests.AIChatRequest) (*responses.AIChatResponse, error) {
	stream, err := a.StreamChat(ctx, request)
	if err != nil {
		return nil, err
	}

	var reply strings.Builder
	for chunk := range stream {
		if chunk.Err != nil {
			return nil, chunk.Err
		}
		reply.WriteString(chunk.Text)
	}

	return &responses.AIChatResponse{Message: reply.String()}, nil
}

func (a *aiUsecase) StreamChat(ctx context.Context, request *requests.AIChatRequest) (<-chan openai_client.StreamChunk, error) {
	if request == nil || strings.TrimSpace(request.Message) == "" {
		return nil, apperror.BadRequestErr()
	}

	userID, ok := ctx.Value(enums.UserID).(uint64)
	if !ok || userID == 0 {
		return nil, apperror.UnauthorizedErr()
	}

	isEnabled, customErr := a.checkFeatureFlag(ctx, userID)
	if customErr != nil {
		return nil, customErr
	}
	if !isEnabled {
		return nil, apperror.ForbiddenErr()
	}
	if a.openaiClient == nil {
		return nil, apperror.InternalServerErr()
	}

	return a.openaiClient.Stream(ctx, senseiInstructions, []openai_client.Turn{
		{Role: "user", Content: request.Message},
	})
}

func (a *aiUsecase) checkFeatureFlag(ctx context.Context, userID uint64) (isEnabled bool, customErr *apperror.CustomError) {
	subscription, err := a.subscriptionRepository.FindCurrentByUserID(ctx, userID)
	if err != nil {
		return false, apperror.DatabaseErrorErr()
	}

	if subscription == nil {
		return false, apperror.NotFoundErr()
	}

	isEnabled, customErr = a.checkFeatureFlagSubscriptionPlan(ctx, subscription.PlanID, "ai_chat")
	if customErr != nil {
		return false, customErr
	}

	return isEnabled, nil
}

func (a *aiUsecase) checkFeatureFlagSubscriptionPlan(ctx context.Context, planID uint64, featureFlag string) (isEnabled bool, customErr *apperror.CustomError) {
	plan, err := a.subscriptionPlanRepository.FindByID(ctx, planID)
	if err != nil {
		return false, apperror.DatabaseErrorErr()
	}

	if plan == nil {
		return false, apperror.NotFoundErr()
	}

	featureFlags := utils.FeatureFlagsToMap(plan.FeatureFlags)
	return featureFlags[featureFlag], nil
}

// GetChat implements [AIUsecase].
func (a *aiUsecase) GetChat(ctx context.Context, request *requests.GetChatRequest) (*responses.ListChatMessagesResponse, error) {
	return nil, apperror.NewCustomError(apperror.ErrInternalServer, "chat history is not implemented", nil)
}

// ListChatRooms implements [AIUsecase].
func (a *aiUsecase) ListChatRooms(ctx context.Context) (*responses.ListChatRoomsResponse, error) {
	return nil, apperror.NewCustomError(apperror.ErrInternalServer, "chat rooms are not implemented", nil)
}

// ListModels implements [AIUsecase].
func (a *aiUsecase) ListModels(ctx context.Context) (*responses.ListModelsResponse, error) {
	return &responses.ListModelsResponse{
		Models: []responses.Model{{ID: "gpt-4o", Name: "gpt-4o"}},
	}, nil
}

func newAIUsecase(
	subscriptionRepository repositories.SubscriptionRepository,
	subscriptionPlanRepository repositories.SubscriptionPlanRepository,
	openaiClient openai_client.OpenAIClient,
) AIUsecase {
	return &aiUsecase{
		subscriptionRepository:     subscriptionRepository,
		subscriptionPlanRepository: subscriptionPlanRepository,
		openaiClient:               openaiClient,
	}
}
