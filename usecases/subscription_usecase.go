package usecases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/caches"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
)

type (
	SubscriptionUsecase interface {
		GetUserSubscriptionDetail(ctx context.Context) (response *responses.GetUserSubscriptionDetailResponse, customErr *apperror.CustomError)
	}

	subscriptionUsecase struct {
		subscriptionRepository     repositories.SubscriptionRepository
		subscriptionCache          caches.SubscriptionCache
		subscriptionPlanRepository repositories.SubscriptionPlanRepository
		userRepository             repositories.UserRepository
	}
)

func (s *subscriptionUsecase) GetUserSubscriptionDetail(ctx context.Context) (response *responses.GetUserSubscriptionDetailResponse, customErr *apperror.CustomError) {
	userID, ok := ctx.Value(enums.UserID).(uint64)
	if !ok || userID == 0 {
		return nil, apperror.NewCustomError(apperror.ErrUnauthorized, "user is missing", fmt.Errorf("user id is missing"))
	}

	userKey := strconv.FormatUint(userID, 10)
	if response, err := s.subscriptionCache.GetUserSubscriptionDetail(ctx, userKey); err == nil {
		return response, nil
	}

	subscriptionModel, err := s.subscriptionRepository.FindCurrentByUserID(ctx, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find subscription", err)
	}
	if subscriptionModel == nil {
		return nil, apperror.NewCustomError(apperror.ErrNotFound, "subscription not found", fmt.Errorf("subscription not found"))
	}

	planModel, err := s.subscriptionPlanRepository.FindByID(ctx, subscriptionModel.PlanID)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find plan", err)
	}
	if planModel == nil {
		return nil, apperror.NewCustomError(apperror.ErrNotFound, "plan not found", fmt.Errorf("plan not found"))
	}

	response = buildGetUserSubscriptionDetailResponse(subscriptionModel, planModel)
	err = s.subscriptionCache.SetUserSubscriptionDetail(ctx, userKey, response)
	if err != nil {
		log.Println("failed to set subscription detail", err)
	}

	return response, nil
}

func buildGetUserSubscriptionDetailResponse(subscriptionModel *models.Subscription, planModel *models.SubscriptionPlan) *responses.GetUserSubscriptionDetailResponse {
	return &responses.GetUserSubscriptionDetailResponse{
		SubscriptionID: subscriptionModel.ID,
		PlanID:         subscriptionModel.PlanID,
		PlanName:       planModel.Name,
		Status:         subscriptionModel.Status,
		CreatedAt:      subscriptionModel.CreatedAt,
		UpdatedAt:      subscriptionModel.UpdatedAt,
		EndDate:        subscriptionModel.EndDate,
		StartDate:      subscriptionModel.StartDate,
	}
}

func newSubscriptionUsecase(
	subscriptionRepository repositories.SubscriptionRepository,
	subscriptionCache caches.SubscriptionCache,
	subscriptionPlanRepository repositories.SubscriptionPlanRepository,
	userRepository repositories.UserRepository,
) SubscriptionUsecase {
	return &subscriptionUsecase{
		subscriptionRepository:     subscriptionRepository,
		subscriptionCache:          subscriptionCache,
		subscriptionPlanRepository: subscriptionPlanRepository,
		userRepository:             userRepository,
	}
}
