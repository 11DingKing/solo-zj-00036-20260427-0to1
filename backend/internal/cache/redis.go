package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"survey-platform/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func InitRedis(cfg *config.RedisConfig) error {
	Redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("unable to connect to redis: %w", err)
	}

	return nil
}

func CloseRedis() {
	if Redis != nil {
		Redis.Close()
	}
}

func Get(ctx context.Context, key string, dest interface{}) error {
	data, err := Redis.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return Redis.Set(ctx, key, data, expiration).Err()
}

func Delete(ctx context.Context, key string) error {
	return Redis.Del(ctx, key).Err()
}

func DeleteByPattern(ctx context.Context, pattern string) error {
	iter := Redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := Redis.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
