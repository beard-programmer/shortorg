package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	DB   int    `json:"db"`
}

func ConnectToRedis(ctx context.Context, logger *zap.SugaredLogger, env string) (*redis.Client, error) {
	config, err := newRedisConfig(env)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:           fmt.Sprintf("%s:%d", config.Host, config.Port),
		DB:             config.DB,
		MaxActiveConns: 100,
	})

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	logger.Infow("Successfully connected to Redis", rdb.PoolStats())

	return rdb, nil
}

func newRedisConfig(env string) (*RedisConfig, error) {
	jsonFile, err := os.ReadFile("./config/redis.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config map[string]RedisConfig
	if err := json.Unmarshal(jsonFile, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	envConfig, exists := config[env]
	if !exists {
		return nil, fmt.Errorf("environment %s not found in config", env)
	}

	return &envConfig, nil
}
