package usecases

import (
	nsq_client "github.com/RandySteven/go-cook/nsq"
	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/caches"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
)

type Usecases struct {
	UserUsecase       UserUsecase
	OnboardingUsecase OnboardingUsecase
}

func NewUsecases(repositories *repositories.Repositories,
	redis *caches.Caches,
	nsq nsq_client.Nsq,
	temporal temporal_client.Temporal,
) *Usecases {
	return &Usecases{
		OnboardingUsecase: newOnboardingUsecase(
			repositories.OnboardingRepository,
			repositories.UserRepository,
			redis.OnboardingCache,
			temporal,
		),
	}
}
