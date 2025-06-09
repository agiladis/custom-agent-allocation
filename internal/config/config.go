package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	AppEnv string `env:"APP_ENV" default:"development`
	Port   string `env:"PORT" default:"8080"`

	QiscusAppID     string `env:"QISCUS_APP_ID" required:"true"`
	QiscusSecretKey string `env:"QISCUSS_SECRET_KEY" required:"true"`
	QiscusBaseURL   string `env:"QISCUS_BASE_URL"`

	DBHost     string `env:"DB_HOST" default:"localhost"`
	DBPort     string `env:"DB_PORT" default:"5432"`
	DBUsername string `env:"DB_USERNAME"`
	DBPassword string `env:"DB_PASSWORD"`
	DBName     string `env:"DB_NAME"`
	DBSSLMode  string `env:"DB_SSLMODE" default:"disable"`

	RedisHost     string `env:"REDIS_HOST" default:"localhost"`
	RedisPort     string `env:"REDIS_PORT" default:"6379"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisStream   string `env:"REDIS_STREAM_NAME" default:"agent_stream"`
	RedisGroup    string `env:"REDIS_CONSUMER_GROUP" default:"agent_consumer_group"`
	RedisConsumer string `env:"REDIS_CONSUMER_NAME" default:"agent-consumer-1"`
}

func Load() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)

	return &cfg, err
}
