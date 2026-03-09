package cache

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// Wrapper cho Redis Client để implement PubSubClient interface
type RedisClientWrapper struct {
	client *redis.Client
}

func NewRedisClient() *RedisClientWrapper {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	// Test connection
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("❌ Failed to connect to Redis:", err)
	}

	log.Println("✅ Connected to Redis")
	return &RedisClientWrapper{client: client}
}

// Close delegates to the underlying redis client
func (r *RedisClientWrapper) Close() error {
	return r.client.Close()
}

// Incr delegates to the underlying redis client (for Rate Limiter)
func (r *RedisClientWrapper) Incr(ctx context.Context, key string) *redis.IntCmd {
	return r.client.Incr(ctx, key)
}

// Expire delegates to the underlying redis client (for Rate Limiter)
func (r *RedisClientWrapper) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return r.client.Expire(ctx, key, expiration)
}

// SubscribeJobStatus implements ports.PubSubClient
func (r *RedisClientWrapper) SubscribeJobStatus(ctx context.Context, jobID string) (<-chan string, error) {
	channelName := "job_status:" + jobID
	pubsub := r.client.Subscribe(ctx, channelName)

	// Verify subscription is successful
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, err
	}

	msgChan := make(chan string)

	go func() {
		defer pubsub.Close()
		defer close(msgChan)

		// Wait for exactly 1 message
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("❌ Redis PubSub error for %s: %v", channelName, err)
			return
		}

		select {
		case msgChan <- msg.Payload:
		case <-ctx.Done():
		}
	}()

	return msgChan, nil
}
