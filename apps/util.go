package apps

import (
	"net"

	db_client "github.com/RandySteven/go-cook/db"
	nsq_client "github.com/RandySteven/go-cook/nsq"
	redis_client "github.com/RandySteven/go-cook/redis"
	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/configs"
	"github.com/RandySteven/jlpt-exercise-be/consumers"
	deepseek_client "github.com/RandySteven/jlpt-exercise-be/pkg/deepseek"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
	oauth2_client "github.com/RandySteven/jlpt-exercise-be/pkg/oauth2"
	smtp_client "github.com/RandySteven/jlpt-exercise-be/pkg/smtp"
)

func prepareDBConfig(config *configs.Config) *db_client.DBConfig {
	dbConfig := config.Configs.Db
	host := dbConfig.Host
	if dbConfig.Port != "" {
		host = net.JoinHostPort(dbConfig.Host, dbConfig.Port)
	}

	return &db_client.DBConfig{
		Db:              dbConfig.Db,
		DbUser:          dbConfig.DbUser,
		DbPass:          dbConfig.DbPass,
		DbHost:          host,
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

func prepareDeepSeekConfig(config *configs.Config) *deepseek_client.DeepSeekConfig {
	deepSeekConfig := config.Configs.DeepSeek

	return &deepseek_client.DeepSeekConfig{
		APIKey: deepSeekConfig.APIKey,
	}
}

func prepareMidtransClient(config *configs.Config) (midtrans_client.MidtransClient, error) {
	midtransConfig := config.Configs.Midtrans
	if midtransConfig.ServerKey == "" {
		return nil, nil
	}
	midtransConfigClient := &midtrans_client.MidtransConfig{
		ServerKey:   midtransConfig.ServerKey,
		ClientKey:   midtransConfig.ClientKey,
		Environment: midtransConfig.Environment,
	}
	return midtrans_client.NewMidtransClient(midtransConfigClient)
}

func prepareSMTPClient(config *configs.Config) (smtp_client.SmtpClient, error) {
	smtpConfig := config.Configs.SMTP
	if smtpConfig.Host == "" || smtpConfig.Port == "" {
		return nil, nil
	}
	return smtp_client.NewSmtpClient(smtpConfig.Host, smtpConfig.Port, smtpConfig.Username, smtpConfig.Password, smtpConfig.From)
}

func prepareOAuth2Config(config *configs.Config) *oauth2_client.OAuth2Config {
	oauth2Config := config.Configs.OAuth2

	return &oauth2_client.OAuth2Config{
		ClientID:     oauth2Config.Google.ClientID,
		ClientSecret: oauth2Config.Google.ClientSecret,
		RedirectURL:  oauth2Config.Google.RedirectURL,
		Scopes:       oauth2Config.Google.Scopes,
	}
}

func (a *App) registerConsumers(runners *consumers.Runners, consumerFunc *consumers.Consumers) {
	runners.RegisterConsumer(`dummy-topic`, consumerFunc.DummyConsumer.Consumer)
	runners.RegisterConsumer(`exercise-topic`, consumerFunc.ExerciseConsumer.Consumer)
}
