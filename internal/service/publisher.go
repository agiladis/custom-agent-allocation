package service

import (
	"context"
	"fmt"

	"github.com/agiladis/custom-agent-allocation/internal/config"
	"github.com/redis/go-redis/v9"
)

type Publisher interface {
	Publish(ctx context.Context, RoomID string) error
}

type redisPublisher struct {
	rdb        *redis.Client
	streamName string
	groupName  string
}

func NewPublisher(ctx context.Context, rdb *redis.Client, cfg *config.Config) (Publisher, error) {
	err := rdb.XGroupCreateMkStream(ctx, cfg.RedisStream, cfg.RedisGroup, "0").Err()
	if err != nil && !isGroupExistsErr(err) {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &redisPublisher{
		rdb:        rdb,
		streamName: cfg.RedisStream,
		groupName:  cfg.RedisGroup,
	}, nil
}

func (p *redisPublisher) Publish(ctx context.Context, roomID string) error {
	if roomID == "" {
		return fmt.Errorf("roomID cannot be empty")
	}

	_, err := p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: p.streamName,
		Values: map[string]interface{}{"room_id": roomID},
	}).Result()

	return err
}

func isGroupExistsErr(err error) bool {
	return err != nil && err.Error() == "BUSYGROUP Consumer Group name already exists"
}
