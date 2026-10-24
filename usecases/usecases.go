package usecases

import (
	nsq_client "github.com/RandySteven/go-cook/nsq"
	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/caches"
	"github.com/RandySteven/jlpt-exercise-be/logic/payment_workflow"
	deepseek_client "github.com/RandySteven/jlpt-exercise-be/pkg/deepseek"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
	oauth2_client "github.com/RandySteven/jlpt-exercise-be/pkg/oauth2"
	smtp_client "github.com/RandySteven/jlpt-exercise-be/pkg/smtp"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/topics"
)

type Usecases struct {
	UserUsecase       UserUsecase
	OnboardingUsecase OnboardingUsecase
	ExerciseUsecase   ExerciseUsecase
	PaymentUsecase    PaymentUsecase
}

func NewUsecases(repositories *repositories.Repositories,
	redis *caches.Caches,
	nsq nsq_client.Nsq,
	temporal temporal_client.Temporal,
	deepSeek deepseek_client.DeepSeek,
	oauth2Client oauth2_client.OAuth2Client,
	midtransClient midtrans_client.MidtransClient,
	smtpClient smtp_client.SmtpClient,
) *Usecases {
	paymentTopic := topics.NewPaymentTopic(nsq)
	paymentWorkflow := payment_workflow.NewPaymentWorkflow(
		temporal_client.NewWorkflowExecution(temporal),
		temporal,
		repositories.PaymentRepository,
		repositories.SubscriptionRepository,
		repositories.SubscriptionPlanRepository,
		repositories.UserRepository,
		repositories.OnboardingRepository,
		midtransClient,
		paymentTopic,
		smtpClient,
	)

	return &Usecases{
		OnboardingUsecase: newOnboardingUsecase(
			repositories.OnboardingRepository,
			repositories.UserRepository,
			redis.OnboardingCache,
			temporal,
			oauth2Client,
			smtpClient,
		),
		ExerciseUsecase: newExerciseUsecase(
			deepSeek,
			temporal,
			repositories.ExerciseRepository,
			repositories.UserRepository,
			repositories.SubscriptionRepository,
			redis.ExerciseCache,
			topics.NewExerciseTopic(nsq),
		),
		PaymentUsecase: newPaymentUsecase(
			paymentWorkflow,
			repositories.SubscriptionRepository,
			repositories.PaymentRepository,
			repositories.UserRepository,
			repositories.OnboardingRepository,
			repositories.SubscriptionPlanRepository,
			midtransClient,
			paymentTopic,
			smtpClient,
			temporal,
		),
	}
}
