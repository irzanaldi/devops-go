package config

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg *Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       0,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis not connected: %v (caching disabled)", err)
		return client
	}

	log.Println("Redis connected successfully")
	return client
}
