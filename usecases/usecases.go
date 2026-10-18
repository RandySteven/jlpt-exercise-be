package usecases

import (
	nsq_client "github.com/RandySteven/go-cook/nsq"
	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/caches"
	deepseek_client "github.com/RandySteven/jlpt-exercise-be/pkg/deepseek"
	oauth2_client "github.com/RandySteven/jlpt-exercise-be/pkg/oauth2"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/topics"
)

type Usecases struct {
	UserUsecase       UserUsecase
	OnboardingUsecase OnboardingUsecase
	ExerciseUsecase   ExerciseUsecase
}

func NewUsecases(repositories *repositories.Repositories,
	redis *caches.Caches,
	nsq nsq_client.Nsq,
	temporal temporal_client.Temporal,
	deepSeek deepseek_client.DeepSeek,
	oauth2Client oauth2_client.OAuth2Client,
) *Usecases {
	return &Usecases{
		OnboardingUsecase: newOnboardingUsecase(
			repositories.OnboardingRepository,
			repositories.UserRepository,
			redis.OnboardingCache,
			temporal,
			oauth2Client,
		),
		ExerciseUsecase: newExerciseUsecase(
			deepSeek,
			temporal,
			repositories.ExerciseRepository,
			repositories.UserRepository,
			redis.ExerciseCache,
			topics.NewExerciseTopic(nsq),
		),
	}
}
