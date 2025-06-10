package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/agiladis/custom-agent-allocation/internal/config"
	v1 "github.com/agiladis/custom-agent-allocation/internal/handler/v1"
	"github.com/agiladis/custom-agent-allocation/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// setup logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// connect to postgres
	// dsn := "host=" + cfg.DBHost +
	// 	" user=" + cfg.DBUsername +
	// 	" password=" + cfg.DBPassword +
	// 	" dbname=" + cfg.DBName +
	// 	" port=" + cfg.DBPort +
	// 	" sslmode=" + cfg.DBSSLMode
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("failed to connect to Postgres")
	// }

	// connect to redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Redis")
	}

	// init fiber
	app := fiber.New()

	// init repo, service, handler
	ctx := context.Background()
	publisher, err := service.NewPublisher(ctx, rdb, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create publisher")
	}
	webhookHandler := v1.NewWebhookHandler(publisher)

	// Routes
	v1 := app.Group("/api/v1")
	v1.Post("/webhook/custom-agent-allocation", webhookHandler.Receive)

	// graceful shutdown
	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal().Err(err).Msg("fiber shutdown failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Info().Msg("shutting down app...")
	_ = app.Shutdown()
}
