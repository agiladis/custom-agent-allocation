package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agiladis/custom-agent-allocation/internal/config"
	"github.com/agiladis/custom-agent-allocation/internal/consumer"
	v1 "github.com/agiladis/custom-agent-allocation/internal/handler/v1"
	"github.com/agiladis/custom-agent-allocation/internal/model"
	"github.com/agiladis/custom-agent-allocation/internal/qiscus"
	"github.com/agiladis/custom-agent-allocation/internal/repository"
	"github.com/agiladis/custom-agent-allocation/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	dsn := cfg.BuildDatabaseDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Postgres")
	}

	// auto migrate
	if err := db.AutoMigrate(&model.AppConfig{}); err != nil {
		log.Fatal().Err(err).Msg("failed to run auto-migrate AppConfig")
	}
	seedAppConfig(db)

	// connect to redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Redis")
	}

	// init repo, service, handler, consumer
	ctx := context.Background()
	cfgRepo := repository.NewConfigRepository(db, rdb)
	cfgSvc := service.NewConfigService(cfgRepo)
	cfgHandler := v1.NewConfigHandler(cfgSvc)

	pub, err := service.NewPublisher(ctx, rdb, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create publisher")
	}
	webhookHandler := v1.NewWebhookHandler(pub)

	qsClient := qiscus.NewClient(cfg)
	assignSvc := service.NewAssignService(cfg, rdb, cfgRepo, qsClient)
	go consumer.RunConsumer(ctx, cfg, rdb, assignSvc) // start consumer in background

	// Fiber and Routes
	app := fiber.New()
	v1 := app.Group("/api/v1")
	v1.Get("/config/max-load", cfgHandler.GetMaxLoad)
	v1.Patch("/config/max-load", cfgHandler.UpdateMaxLoad)
	v1.Post("/webhook/custom-agent-allocation", webhookHandler.Receive)

	// start server with graceful shutdown
	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal().Err(err).Msg("Server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Info().Msg("shutting down app...")
	_ = app.Shutdown()
}

func seedAppConfig(db *gorm.DB) {
	const (
		configKey    = "max_load"
		defaultValue = "2"
	)

	var c model.AppConfig
	err := db.First(&c, "key = ?", configKey).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c = model.AppConfig{
			Key:       configKey,
			Value:     defaultValue,
			UpdatedAt: time.Now(),
		}
		if err := db.Create(&c).Error; err != nil {
			log.Fatal().Err(err).Msg("failed to seed max_load")
		}
		return
	}
}
