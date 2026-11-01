package usecases

import (
	"context"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/logic/ai_chat_workflow"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/utils"
)

type (
	AIUsecase interface {
		Chat(ctx context.Context, request *requests.AIChatRequest) (*responses.AIChatResponse, error)
		ListModels(ctx context.Context) (*responses.ListModelsResponse, error)
		ListChatRooms(ctx context.Context) (*responses.ListChatRoomsResponse, error)
		GetChat(ctx context.Context, request *requests.GetChatRequest) error
	}

	aiUsecase struct {
		subscriptionRepository     repositories.SubscriptionRepository
		subscriptionPlanRepository repositories.SubscriptionPlanRepository
		aiChatWorkflow             ai_chat_workflow.AIChatWorkflow
	}
)

// Chat implements [AIUsecase].
func (a *aiUsecase) Chat(ctx context.Context, request *requests.AIChatRequest) (*responses.AIChatResponse, error) {
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

	a.aiChatWorkflow.Chat(ctx, request)

	return nil, nil
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
func (a *aiUsecase) GetChat(ctx context.Context, request *requests.GetChatRequest) error {
	panic("unimplemented")
}

// ListChatRooms implements [AIUsecase].
func (a *aiUsecase) ListChatRooms(ctx context.Context) (*responses.ListChatRoomsResponse, error) {
	panic("unimplemented")
}

// ListModels implements [AIUsecase].
func (a *aiUsecase) ListModels(ctx context.Context) (*responses.ListModelsResponse, error) {
	panic("unimplemented")
}

func newAIUsecase(subscriptionRepository repositories.SubscriptionRepository) AIUsecase {
	return &aiUsecase{
		subscriptionRepository: subscriptionRepository,
	}
}
