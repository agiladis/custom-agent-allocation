package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppEnv string `envconfig:"APP_ENV" default:"development"`
	Port   string `envconfig:"PORT" default:"8080"`

	QiscusAppID     string `envconfig:"QISCUS_APP_ID" required:"true"`
	QiscusSecretKey string `envconfig:"QISCUS_SECRET_KEY" required:"true"`
	QiscusBaseURL   string `envconfig:"QISCUS_BASE_URL"`

	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     string `envconfig:"DB_PORT" default:"5432"`
	DBUsername string `envconfig:"DB_USERNAME"`
	DBPassword string `envconfig:"DB_PASSWORD"`
	DBName     string `envconfig:"DB_NAME"`
	DBSSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`

	RedisHost     string `envconfig:"REDIS_HOST" default:"localhost"`
	RedisPort     string `envconfig:"REDIS_PORT" default:"6379"`
	RedisPassword string `envconfig:"REDIS_PASSWORD"`
	RedisStream   string `envconfig:"REDIS_STREAM_NAME" default:"agent_stream"`
	RedisGroup    string `envconfig:"REDIS_CONSUMER_GROUP" default:"agent_consumer_group"`
	RedisConsumer string `envconfig:"REDIS_CONSUMER_NAME" default:"agent-consumer-1"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
