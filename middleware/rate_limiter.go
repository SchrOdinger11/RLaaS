package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/SchrOdinger11/RLaaS/redis"
)

// RateLimiter holds our rate limiting configuration and Redis client.
type RateLimiter struct {
	RedisClient *redis.Client
	Limit       int           // Maximum allowed requests per window
	Window      time.Duration // Duration of the window (e.g., 1 minute)
}

// Allow checks if the client identified by apiKey is allowed to make a request.
// It implements a fixed window counter using Redis.
func (rl *RateLimiter) Allow(apiKey string) (bool, error) {
	ctx := context.Background()
	now := time.Now().Unix()
	windowStart := now - (now % int64(rl.Window.Seconds()))
	key := fmt.Sprintf("rate:%s:%d", apiKey, windowStart)

	// Get current request count from Redis.
	count, err := rl.RedisClient.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, err
	}

	// If count exceeds limit, block the request.
	if count >= rl.Limit {
		return false, nil
	}

	// Use a Redis transaction (pipeline) to increment and set expiration.
	pipe := rl.RedisClient.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, rl.Window)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Middleware applies the rate limiting check to every HTTP request.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve API key from request header.
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			http.Error(w, "API key required", http.StatusUnauthorized)
			return
		}
		// Check if the request is allowed.
		allowed, err := rl.Allow(apiKey)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if !allowed {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		// Proceed to the next handler if allowed.
		next.ServeHTTP(w, r)
	})
}
