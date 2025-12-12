package middleware

import (
	"compost-bin/logger"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var cache *redis.Client

func GetCache() *redis.Client {
	return cache
}

func init() {
	addr := getEnv("REDIS_HOST", "127.0.0.1") + ":" + getEnv("REDIS_PORT", "6379")
	password := getEnv("REDIS_PASSWORD", "password")

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		logger.Fatalf("Failed to connect to redis: %v", err)
	}
	cache = client
}
