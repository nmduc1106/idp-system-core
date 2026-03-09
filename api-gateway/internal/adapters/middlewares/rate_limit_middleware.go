package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RedisClient defines the interface required by RateLimitMiddleware
type RedisClient interface {
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}

// RateLimitMiddleware applies rate limiting using a Redis counter.
func RateLimitMiddleware(redisClient RedisClient, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract userID from context (set by JWTMiddleware)
		userID, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing user ID for rate limit"})
			return
		}

		key := fmt.Sprintf("rate_limit:upload:%v", userID)
		ctx := context.Background()

		// Increment the counter
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error: Rate limiter failure"})
			return
		}

		// Set expiration on the first increment
		if count == 1 {
			redisClient.Expire(ctx, key, window)
		}

		// Check rate limit
		if count > int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too Many Requests"})
			return
		}

		c.Next()
	}
}
