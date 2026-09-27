package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/irzan-aldi/task-manager/pkg/response"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(redisClient *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", ip)

		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			// Redis down — allow request
			c.Next()
			return
		}

		if count == 1 {
			redisClient.Expire(ctx, key, window)
		}

		if count > int64(limit) {
			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded. Try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}
