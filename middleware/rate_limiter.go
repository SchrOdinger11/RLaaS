package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/SchrOdinger11/RLaaS/redis"
)

// RateLimiter holds our rate limiting configuration and Redis client.
// DefaultLimit and DefaultWindow are used if no dynamic configuration is found.
type RateLimiter struct {
	RedisClient   *redis.Client
	DefaultLimit  int           // Default maximum allowed requests per window
	DefaultWindow time.Duration // Default duration of the window (e.g., 1 minute)
}

// fetchConfig checks Redis for a per-API-key configuration.
// It returns limit and window (in seconds). If none is found, it returns default values.
func (rl *RateLimiter) fetchConfig(apiKey string) (int, int, error) {
	ctx := context.Background()
	key := "config:" + apiKey
	configMap, err := rl.RedisClient.HGetAll(ctx, key).Result()
	if err != nil {
		return rl.DefaultLimit, int(rl.DefaultWindow.Seconds()), err
	}
	// If no configuration is set, return defaults.
	if len(configMap) == 0 {
		return rl.DefaultLimit, int(rl.DefaultWindow.Seconds()), nil
	}
	limit, err := strconv.Atoi(configMap["limit"])
	if err != nil {
		return rl.DefaultLimit, int(rl.DefaultWindow.Seconds()), err
	}
	window, err := strconv.Atoi(configMap["window"])
	if err != nil {
		return rl.DefaultLimit, int(rl.DefaultWindow.Seconds()), err
	}
	return limit, window, nil
}

// Allow checks if the client identified by apiKey is allowed to make a request.
// It uses dynamic configuration if available, otherwise falls back to defaults.
func (rl *RateLimiter) Allow(apiKey string) (bool, error) {
	ctx := context.Background()
	// Get dynamic configuration for the API key.
	limit, windowSeconds, _ := rl.fetchConfig(apiKey)
	now := time.Now().Unix()
	windowStart := now - (now % int64(windowSeconds))
	key := fmt.Sprintf("rate:%s:%d", apiKey, windowStart)

	// Get current request count from Redis.
	count, err := rl.RedisClient.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, err
	}

	// If count exceeds limit, block the request.
	if count >= limit {
		return false, nil
	}

	// Use a Redis transaction (pipeline) to increment and set expiration.
	pipe := rl.RedisClient.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)
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
