package apps

import (
	"context"
	"os"

	db_client "github.com/RandySteven/go-cook/db"
	nsq_client "github.com/RandySteven/go-cook/nsq"
	redis_client "github.com/RandySteven/go-cook/redis"
	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/go-kopi/caches"
	"github.com/RandySteven/go-kopi/configs"
	"github.com/RandySteven/go-kopi/consumers"
	rest_handler "github.com/RandySteven/go-kopi/handlers"
	"github.com/RandySteven/go-kopi/repositories"
	"github.com/RandySteven/go-kopi/topics"
	"github.com/RandySteven/go-kopi/usecases"
)

type (
	App struct {
		MySQL    db_client.DBClient
		Redis    redis_client.Redis
		Temporal temporal_client.Temporal
		Nsq      nsq_client.Nsq
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
	return &App{
		MySQL:    mysqlClient,
		Redis:    redisClient,
		Nsq:      nsqClient,
		Temporal: temporalClient,
	}, nil
}

func (a *App) PrepareHttpHandler(ctx context.Context) *rest_handler.Handlers {
	repositories := repositories.NewRepositories(a.MySQL.Client())
	caches := caches.NewCaches(a.Redis.Client())
	topics := topics.NewTopics(a.Nsq)
	usecases := usecases.NewUsecases(repositories, caches, a.Nsq, a.Temporal)
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
	caches := caches.NewCaches(a.Redis.Client())
	consumerFuncs := consumers.NewConsumers(repositories, caches, topics)
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
	migrationWorker := db_client.MigrationWorker{}
	if err := migrationWorker.Migration(ctx); err != nil {
		return err
	}
	return nil
}
