package cache

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

// Client is the shared Redis handle used by the caching repository decorators.
// It stays nil when caching is disabled.
var Client *redis.Client

// Enabled reports whether caching is turned on via the CACHE_ENABLED env var.
// Defaults to off so local/dev usage keeps working without a Redis instance.
func Enabled() bool {
	return getEnv("CACHE_ENABLED", "false") == "true"
}

// InitCache connects to Redis when caching is enabled. No-op otherwise.
func InitCache() {
	if !Enabled() {
		return
	}

	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", "localhost"), getEnv("REDIS_PORT", "6379")),
		Password: getEnv("REDIS_PASSWORD", ""),
	})

	if err := Client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
