package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const cacheKey = "config:max_load"

type ConfigRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewConfigRepository(db *gorm.DB, rdb *redis.Client) *ConfigRepository {
	return &ConfigRepository{db: db, rdb: rdb}
}

func (r *ConfigRepository) GetMaxLoad(ctx context.Context) (int, error) {
	// 1. Get from cache
	if val, err := r.rdb.Get(ctx, cacheKey).Result(); err == nil {
		if i, err := strconv.Atoi(val); err == nil {
			return i, nil
		}
	}

	// 2. fallback: Read  from Postgres
	var row struct {
		Key       string
		Value     string
		UpdatedAt time.Time
	}
	if err := r.db.WithContext(ctx).
		Table("app_config").
		Where("key = ?", "max_load").
		Select("key", "value", "updated_at").
		Take(&row).Error; err != nil {
		return 0, fmt.Errorf("db read max_load: %w", err)
	}

	maxLoad, err := strconv.Atoi(row.Value)
	if err != nil {
		return 0, fmt.Errorf("invalid value in db for max_load: %w", err)
	}

	// Set to Redis cache
	_ = r.rdb.Set(ctx, cacheKey, row.Value, 0).Err()

	return maxLoad, nil
}

func (r *ConfigRepository) UpdateMaxLoad(ctx context.Context, newVal int) error {
	strVal := strconv.Itoa(newVal)

	// 1. Update DB
	if err := r.db.WithContext(ctx).
		Table("app_config").
		Where("key = ?", "max_load").
		UpdateColumns(map[string]interface{}{
			"value":      strVal,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("db update max_load: %w", err)
	}

	// 2. Update Redis cache
	if err := r.rdb.Set(ctx, cacheKey, strVal, 0).Err(); err != nil {
		return fmt.Errorf("redis cache set max_load: %w", err)
	}

	return nil
}
