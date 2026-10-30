package apps

import (
	"context"
	"os"

	db_client "github.com/RandySteven/go-cook/db"
	nsq_client "github.com/RandySteven/go-cook/nsq"
	redis_client "github.com/RandySteven/go-cook/redis"
	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/caches"
	"github.com/RandySteven/jlpt-exercise-be/configs"
	"github.com/RandySteven/jlpt-exercise-be/consumers"
	rest_handler "github.com/RandySteven/jlpt-exercise-be/handlers"
	deepseek_client "github.com/RandySteven/jlpt-exercise-be/pkg/deepseek"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
	oauth2_client "github.com/RandySteven/jlpt-exercise-be/pkg/oauth2"
	pdf_client "github.com/RandySteven/jlpt-exercise-be/pkg/pdf"
	smtp_client "github.com/RandySteven/jlpt-exercise-be/pkg/smtp"
	"github.com/RandySteven/jlpt-exercise-be/queries"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/topics"
	"github.com/RandySteven/jlpt-exercise-be/usecases"
)

type (
	App struct {
		MySQL    db_client.DBClient
		Redis    redis_client.Redis
		Temporal temporal_client.Temporal
		Nsq      nsq_client.Nsq
		DeepSeek deepseek_client.DeepSeek
		OAuth2   oauth2_client.OAuth2Client
		Midtrans midtrans_client.MidtransClient
		Smtp     smtp_client.SmtpClient
		PDF      pdf_client.PDFClient
	}
)

func NewApp(config *configs.Config) (*App, error) {
	mysqlClient, err := db_client.NewMYSQLClient(prepareDBConfig(config))
	if err != nil {
		return nil, err
	}
	nsqClient, err := nsq_client.NewNsqClient(prepareNSQConfig(config))
	if err != nil {
		return nil, err
	}
	redisClient, err := redis_client.NewRedisClient(prepareRedisConfig(config))
	if err != nil {
		return nil, err
	}
	temporalClient, err := temporal_client.NewTemporalClient(prepareTemporalConfig(config))
	if err != nil {
		return nil, err
	}
	deepseekClient, err := deepseek_client.NewDeepSeek(prepareDeepSeekConfig(config))
	if err != nil {
		return nil, err
	}
	oauth2Client, err := oauth2_client.NewOAuth2Client(prepareOAuth2Config(config))
	if err != nil {
		return nil, err
	}
	midtransClient, err := prepareMidtransClient(config)
	if err != nil {
		return nil, err
	}
	smtpClient, err := prepareSMTPClient(config)
	if err != nil {
		return nil, err
	}
	pdfClient, err := pdf_client.NewPDFClient()
	if err != nil {
		return nil, err
	}
	return &App{
		MySQL:    mysqlClient,
		Redis:    redisClient,
		Nsq:      nsqClient,
		Temporal: temporalClient,
		DeepSeek: deepseekClient,
		OAuth2:   oauth2Client,
		Midtrans: midtransClient,
		Smtp:     smtpClient,
		PDF:      pdfClient,
	}, nil
}

func (a *App) PrepareHttpHandler(ctx context.Context) *rest_handler.Handlers {
	repositories := repositories.NewRepositories(a.MySQL.Client())
	caches := caches.NewCaches(a.Redis)
	topics := topics.NewTopics(a.Nsq)
	usecases := usecases.NewUsecases(repositories, caches, a.Nsq, a.Temporal, a.DeepSeek, a.OAuth2, a.Midtrans, a.Smtp, a.PDF)
	return rest_handler.NewHandlers(usecases, topics)
}

func (a *App) RefreshRedis(ctx context.Context) error {
	return a.Redis.ClearCache(ctx)
}

func (a *App) PrepareJobScheduler(ctx context.Context) {

}

func (a *App) PrepareConsumer(ctx context.Context) *consumers.Runners {
	topics := topics.NewTopics(a.Nsq)
	repositories := repositories.NewRepositories(a.MySQL.Client())
	caches := caches.NewCaches(a.Redis)
	consumerFuncs := consumers.NewConsumers(repositories, caches, topics, a.DeepSeek)
	channelName := os.Getenv(`NSQ_CHANNEL`)
	if channelName == `` {
		channelName = `default`
	}
	consumerRunner := consumers.InitRunner(a.Nsq, channelName)
	a.registerConsumers(consumerRunner, consumerFuncs)
	return consumerRunner
}

func (a *App) ExecuteMigration(ctx context.Context) error {
	defer a.MySQL.Close()
	migrationWorker := db_client.InitMigrationWorker(a.MySQL)
	migrationWorker.RegisterMigration(
		queries.MigrateOnboarding.ToString(),
		queries.MigrateOnboardingRegisterAsIndex.ToString(),
		queries.MigrateOnboardingStatusIndex.ToString(),
		queries.MigrateUser.ToString(),
		queries.MigrateUserStatusIndex.ToString(),
		queries.MigrateExercise.ToString(),
		queries.MigrateExerciseReferenceID.ToString(),
		queries.MigrateExerciseAnswer.ToString(),
		queries.MigrateExerciseUserIndex.ToString(),
		queries.MigrateExerciseStatusIndex.ToString(),
		queries.MigrateExerciseNLevelIndex.ToString(),
		queries.DropExerciseReferenceIDIndex.ToString(),
		// queries.MigrateExerciseReferenceIDIndex.ToString(),
		// queries.MigrateExerciseOneOpenPerUserIndex.ToString(),
		queries.MigrateSubscriptionPlan.ToString(),
		queries.MigrateSubscriptionPlanColumns.ToString(),
		queries.MigrateSubscriptionPlanNameIndex.ToString(),
		queries.SeedSubscriptionPlans.ToString(),
		queries.MigrateSubscription.ToString(),
		queries.MigrateSubscriptionUserIndex.ToString(),
		queries.MigrateSubscriptionPlanIndex.ToString(),
		queries.MigrateSubscriptionStatusIndex.ToString(),
		queries.MigrateSubscriptionReferenceID.ToString(),
		queries.MigrateSubscriptionReferenceIDIndex.ToString(),
		queries.MigratePayment.ToString(),
		queries.MigratePaymentUserIndex.ToString(),
		queries.MigratePaymentSubscriptionIndex.ToString(),
		queries.MigratePaymentStatusIndex.ToString(),
		queries.MigratePaymentReferenceIDIndex.ToString(),
	)
	if err := migrationWorker.Migration(ctx); err != nil {
		return err
	}
	return nil
}
