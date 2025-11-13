package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"go.uber.org/zap"
)

func SetupRedis(ctx context.Context, config *Config, logger *zap.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("Redis connection established successfully")

	return client, nil
}

// for security reasons, we will set the initial counter value to 14000000 - 1
// and then increment it by 1 each time a new short URL is created. This is to prevent
// the counter from being guessed by the public.
func SetInitialCounterValue(ctx context.Context, client *redis.Client, config *Config, logger *zap.Logger) (bool, error) {
	initialValue := int64(config.CounterStartVal - 1)

	set, err := client.SetNX(ctx, config.CounterKey, initialValue, 0).Result()
	if err != nil {
		return false, fmt.Errorf("failed to initialize counter: %w", err)
	}

	if set {
		logger.Info("Initialized Redis counter", zap.String("key", config.CounterKey), zap.Int64("start_value", initialValue))
	}

	return set, nil
}
