package usecases

import (
	"context"
	"log"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/caches"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/utils"
)

type (
	SubscriptionPlanUsecase interface {
		AddSubscriptionPlan(ctx context.Context, request *requests.AddSubscriptionPlanRequest) (response *responses.AddSubscriptionPlanResponse, customErr *apperror.CustomError)
		ListSubscriptionPlans(ctx context.Context) (response *responses.ListSubscriptionPlansResponse, customErr *apperror.CustomError)
		GetSubscriptionPlanDetail(ctx context.Context, id uint64) (response *responses.GetSubscriptionPlanDetailResponse, customErr *apperror.CustomError)
	}

	subscriptionPlanUsecase struct {
		subscriptionPlanRepository repositories.SubscriptionPlanRepository
		subscriptionPlanCache      caches.SubscriptionPlanCache
		singleFlightGroup          utils.SingleFlightGroup
	}
)

// AddSubscriptionPlan implements [SubscriptionPlanUsecase].
func (s *subscriptionPlanUsecase) AddSubscriptionPlan(ctx context.Context, request *requests.AddSubscriptionPlanRequest) (response *responses.AddSubscriptionPlanResponse, customErr *apperror.CustomError) {
	panic("unimplemented")
}

// GetSubscriptionPlanDetail implements [SubscriptionPlanUsecase].
func (s *subscriptionPlanUsecase) GetSubscriptionPlanDetail(ctx context.Context, id uint64) (response *responses.GetSubscriptionPlanDetailResponse, customErr *apperror.CustomError) {
	subscriptionID := string(id)
	//singleflight group to get subscription plan detail
	subscriptionResponse, err := s.singleFlightGroup.Do(subscriptionID, func() (interface{}, error) {
		if subscription, err := s.getSubscriptionPlanDetailCache(ctx, id); err == nil {
			if subscription != nil {
				return subscription, nil
			}
		}

		subscription, err := s.getSubscriptionPlanDetailDatabase(ctx, id)
		if err != nil {
			return nil, err
		}

		// set cache
		err = s.subscriptionPlanCache.SetSubscriptionPlanDetail(ctx, id, subscription)
		if err != nil {
			log.Println("failed to set subscription plan detail", err)
		}

		return subscription, nil
	})
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to get subscription plan detail", err)
	}

	response = subscriptionResponse.(*responses.GetSubscriptionPlanDetailResponse)
	return response, nil
}

func (s *subscriptionPlanUsecase) getSubscriptionPlanDetailCache(ctx context.Context, id uint64) (response *responses.GetSubscriptionPlanDetailResponse, err error) {
	response, err = s.subscriptionPlanCache.GetSubscriptionPlanDetail(ctx, id)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to get subscription plan detail", err)
	}
	return response, nil
}

func (s *subscriptionPlanUsecase) getSubscriptionPlanDetailDatabase(ctx context.Context, id uint64) (response *responses.GetSubscriptionPlanDetailResponse, err error) {
	subscriptionPlan, err := s.subscriptionPlanRepository.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to get subscription plan detail", err)
	}
	return &responses.GetSubscriptionPlanDetailResponse{
		ID:   subscriptionPlan.ID,
		Name: subscriptionPlan.Name,
	}, nil
}

// ListSubscriptionPlans implements [SubscriptionPlanUsecase].
func (s *subscriptionPlanUsecase) ListSubscriptionPlans(ctx context.Context) (response *responses.ListSubscriptionPlansResponse, customErr *apperror.CustomError) {
	subscriptionPlans, err := s.subscriptionPlanRepository.FindAll(ctx, 0, 0)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to list subscription plans", err)
	}

	subscriptionPlansResponse := make([]*responses.GetSubscriptionPlanDetailResponse, len(subscriptionPlans))
	for i, subscriptionPlan := range subscriptionPlans {
		subscriptionPlansResponse[i] = &responses.GetSubscriptionPlanDetailResponse{
			ID:   subscriptionPlan.ID,
			Name: subscriptionPlan.Name,
		}
	}

	response = &responses.ListSubscriptionPlansResponse{SubscriptionPlans: subscriptionPlansResponse}
	return response, nil
}

func newSubscriptionPlanUsecase(
	subscriptionPlanRepository repositories.SubscriptionPlanRepository,
	subscriptionPlanCache caches.SubscriptionPlanCache,
) SubscriptionPlanUsecase {
	return &subscriptionPlanUsecase{
		subscriptionPlanRepository: subscriptionPlanRepository,
		subscriptionPlanCache:      subscriptionPlanCache,
	}
}
