// Package config provides configuration management for the application.
// It handles loading, parsing, and accessing configuration values from YAML files
// and supports server, database (PostgreSQL, MongoDB), Redis, NSQ, and Temporal settings.
package configs

import "time"

type (
	// Config represents the root configuration structure for the application.
	// It contains nested configuration for all services including server, databases,
	// message queues, and workflow orchestration.
	Config struct {
		Configs struct {
			// Server contains HTTP server configuration including host, port, and timeouts.
			Server struct {
				Host    string `yaml:"host"`
				Port    string `yaml:"port"`
				Timeout struct {
					Server time.Duration `yaml:"server"`
					Read   time.Duration `yaml:"read"`
					Write  time.Duration `yaml:"write"`
					Idle   time.Duration `yaml:"idle"`
				} `yaml:"timeout"`
			} `yaml:"server"`

			// Postgres contains PostgreSQL database connection settings.
			Db struct {
				Db              string `yaml:"db"`
				Host            string `yaml:"host"`
				Port            string `yaml:"port"`
				DbName          string `yaml:"dbname"`
				DbUser          string `yaml:"dbuser"`
				DbPass          string `yaml:"dbpass"`
				SSLMode         string `yaml:"ssl_mode"`
				MaxIdleConns    int    `yaml:"max_idle_conns"`
				MaxOpenConns    int    `yaml:"max_open_conns"`
				ConnMaxLifeTime int    `yaml:"conn_max_lifetime"`
				ConnMaxIdleTime int    `yaml:"conn_max_idletime"`
			} `yaml:"db"`

			// Mongodb contains MongoDB connection settings.
			Mongodb struct {
				User     string `yaml:"user"`
				Password string `yaml:"password"`
				Host     string `yaml:"host"`
				DbName   string `yaml:"dbname"`
			} `yaml:"mongodb"`

			// Redis contains Redis cache connection settings.
			Redis struct {
				Host            string `yaml:"host"`
				Port            string `yaml:"port"`
				Password        string `yaml:"password"`
				PoolSize        int    `yaml:"pool_size"`
				MinIdleConn     int    `yaml:"min_idle_conn"`
				MaxIdleConn     int    `yaml:"max_idle_conn"`
				PoolTimeout     int    `yaml:"pool_timeout"`
				DialTimeout     int    `yaml:"dial_timeout"`
				ReadTimeout     int    `yaml:"read_timeout"`
				WriteTimeout    int    `yaml:"write_timeout"`
				MaxRetries      int    `yaml:"max_retries"`
				MaxRetryBackoff int    `yaml:"max_retry_backoff"`
				MinRetryBackoff int    `yaml:"min_retry_backoff"`
				ConnMaxIdleTime int    `yaml:"conn_max_idle_time"`
				ConnMaxLifeTime int    `yaml:"conn_max_life_time"`
			} `yaml:"redis"`

			// NSQ contains NSQ message queue settings.
			NSQ struct {
				Host               string `yaml:"host"`
				TCPPort            string `yaml:"tcp_port"`
				LookupdHttpPort    string `yaml:"lookup_http_port"`
				MaxInFlight        int    `yaml:"max_in_flight"`
				ConcurrentConsumer int    `yaml:"concurrent_consumer"`
				ReadTimeout        int    `yaml:"read_timeout"`
				WriteTimeout       int    `yaml:"write_timeout"`
				HeartbeatInterval  int    `yaml:"heartbeat_interval"`
				BackoffMultiplier  int    `yaml:"backoff_multiplier"`
				MaxBackoffDuration int    `yaml:"max_backoff_duration"`
			} `yaml:"nsq"`

			// Temporal contains Temporal workflow orchestration settings.
			Temporal struct {
				Host          string `yaml:"host"`
				Port          string `yaml:"port"`
				TaskQueue     string `yaml:"task_queue"`
				Namespace     string `yaml:"namespace"`
				WorkerOptions *struct {
					MaxConcurrentActivityExecutionSize      int     `yaml:"maxConcurrentActivityExecutionSize"`
					WorkerActivitiesPerSecond               float64 `yaml:"workerActivitiesPerSecond"`
					MaxConcurrentLocalActivityExecutionSize int     `yaml:"maxConcurrentLocalActivityExecutionSize"`
					WorkerLocalActivitiesPerSecond          float64 `yaml:"workerLocalActivitiesPerSecond"`
				} `yaml:"workerOptions"`
			} `yaml:"temporal"`
		} `yaml:"configs"`
	}
)

// GetConfigs returns the nested Configs struct containing all service configurations.
func (c *Config) GetConfigs() any {
	return &c.Configs
}
