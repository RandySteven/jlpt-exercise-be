package usecases

import "context"

type (
	SubscriptionPlanUsecase interface {
		AddSubscriptionPlan(ctx context.Context)
		ListSubscriptionPlans(ctx context.Context)
		GetSubscription(ctx context.Context)
	}
)
