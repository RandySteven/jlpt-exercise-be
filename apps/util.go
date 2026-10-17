package apps

import (
	db_client "github.com/RandySteven/go-cook/db"
	nsq_client "github.com/RandySteven/go-cook/nsq"
	redis_client "github.com/RandySteven/go-cook/redis"
	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/go-kopi/configs"
	"github.com/RandySteven/go-kopi/consumers"
)

func prepareDBConfig(config *configs.Config) *db_client.DBConfig {
	dbConfig := config.Configs.Db

	return &db_client.DBConfig{
		Db:              dbConfig.Db,
		DbUser:          dbConfig.DbUser,
		DbPass:          dbConfig.DbPass,
		DbHost:          dbConfig.Host,
		DbName:          dbConfig.DbName,
		SSLMode:         dbConfig.SSLMode,
		MaxIdleConns:    dbConfig.MaxIdleConns,
		MaxOpenConns:    dbConfig.MaxOpenConns,
		ConnMaxLifeTime: dbConfig.ConnMaxIdleTime,
		ConnMaxIdleTime: dbConfig.ConnMaxIdleTime,
	}
}

func prepareRedisConfig(config *configs.Config) *redis_client.RedisConfig {
	redisConfig := config.Configs.Redis

	return &redis_client.RedisConfig{
		Host:            redisConfig.Host,
		Port:            redisConfig.Port,
		Password:        redisConfig.Password,
		PoolSize:        redisConfig.PoolSize,
		MinIdleConn:     redisConfig.MinIdleConn,
		MaxIdleConn:     redisConfig.MaxIdleConn,
		PoolTimeout:     redisConfig.PoolTimeout,
		DialTimeout:     redisConfig.DialTimeout,
		ReadTimeout:     redisConfig.ReadTimeout,
		WriteTimeout:    redisConfig.WriteTimeout,
		MaxRetries:      redisConfig.MaxRetries,
		MaxRetryBackoff: redisConfig.MaxRetryBackoff,
		MinRetryBackoff: redisConfig.MinRetryBackoff,
		ConnMaxIdleTime: redisConfig.ConnMaxIdleTime,
		ConnMaxLifeTime: redisConfig.ConnMaxLifeTime,
	}
}

func prepareNSQConfig(config *configs.Config) *nsq_client.NSQConfig {
	nsqConfig := config.Configs.NSQ

	return &nsq_client.NSQConfig{
		NSQDHost:           nsqConfig.Host,
		NSQDTCPPort:        nsqConfig.TCPPort,
		LookupdHttpPort:    nsqConfig.LookupdHttpPort,
		MaxInFlight:        nsqConfig.MaxInFlight,
		ConcurrentConsumer: nsqConfig.ConcurrentConsumer,
		ReadTimeout:        nsqConfig.ReadTimeout,
		WriteTimeout:       nsqConfig.WriteTimeout,
		BackoffMultiplier:  nsqConfig.BackoffMultiplier,
		MaxBackoffDuration: nsqConfig.MaxBackoffDuration,
	}
}

func prepareTemporalConfig(config *configs.Config) *temporal_client.TemporalConfig {
	temporalConfig := config.Configs.Temporal

	cfg := &temporal_client.TemporalConfig{
		Host:      temporalConfig.Host,
		Port:      temporalConfig.Port,
		Namespace: temporalConfig.Namespace,
		TaskQueue: temporalConfig.TaskQueue,
	}

	cfg.WorkerOptions.MaxConcurrentActivityExecutionSize = temporalConfig.WorkerOptions.MaxConcurrentActivityExecutionSize
	cfg.WorkerOptions.MaxConcurrentLocalActivityExecutionSize = temporalConfig.WorkerOptions.MaxConcurrentLocalActivityExecutionSize
	cfg.WorkerOptions.WorkerActivitiesPerSecond = temporalConfig.WorkerOptions.WorkerActivitiesPerSecond
	cfg.WorkerOptions.WorkerLocalActivitiesPerSecond = temporalConfig.WorkerOptions.WorkerLocalActivitiesPerSecond

	return cfg
}

func (a *App) registerConsumers(runners *consumers.Runners, consumerFunc *consumers.Consumers) {
	runners.RegisterConsumer(`dummy-topic`, consumerFunc.DummyConsumer.DummyFunc)
}
