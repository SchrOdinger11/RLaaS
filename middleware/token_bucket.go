package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/SchrOdinger11/RLaaS/redis"
)

// TokenBucketLimiter implements a token bucket algorithm.
type TokenBucketLimiter struct {
	RedisClient  *redis.Client
	Capacity     int           // Maximum tokens available.
	RefillRate   int           // Tokens added per second.
	RefillWindow time.Duration // Duration over which tokens are refilled (typically 1 second).
}

// getBucket retrieves the current bucket state from Redis.
// It returns: current tokens, capacity, last refill timestamp.
func (tb *TokenBucketLimiter) getBucket(apiKey string) (int, int, int64, error) {
	ctx := context.Background()
	key := "bucket:" + apiKey
	result, err := tb.RedisClient.HGetAll(ctx, key).Result()
	if err != nil {
		return 0, 0, 0, err
	}
	// If the bucket doesn't exist, initialize with full capacity.
	if len(result) == 0 {
		return tb.Capacity, tb.Capacity, time.Now().Unix(), nil
	}
	tokens, err := strconv.Atoi(result["tokens"])
	if err != nil {
		return 0, 0, 0, err
	}
	capacity, err := strconv.Atoi(result["capacity"])
	if err != nil {
		capacity = tb.Capacity
	}
	lastRefill, err := strconv.ParseInt(result["last_refill"], 10, 64)
	if err != nil {
		lastRefill = time.Now().Unix()
	}
	return tokens, capacity, lastRefill, nil
}

// setBucket updates the bucket state in Redis.
func (tb *TokenBucketLimiter) setBucket(apiKey string, tokens int, lastRefill int64) error {
	ctx := context.Background()
	key := "bucket:" + apiKey
	_, err := tb.RedisClient.HSet(ctx, key, map[string]interface{}{
		"tokens":      tokens,
		"capacity":    tb.Capacity,
		"last_refill": lastRefill,
	}).Result()
	return err
}

// min returns the minimum of two integers.
func min(a, b int) int {

	if a < b {
		return a
	}
	return b
}

// Allow checks whether a request is allowed by consuming a token from the bucket.
func (tb *TokenBucketLimiter) Allow(apiKey string) (bool, error) {

	// Retrieve current bucket state.
	tokens, capacity, lastRefill, err := tb.getBucket(apiKey)
	if err != nil {
		return false, err
	}
	now := time.Now().Unix()

	// Calculate elapsed time since last refill.
	elapsed := now - lastRefill

	// Calculate tokens to add based on elapsed time and refill rate.
	tokensToAdd := int(elapsed) * tb.RefillRate
	if tokensToAdd > 0 {
		tokens = min(tokens+tokensToAdd, capacity)
		lastRefill = now
	}

	// If no tokens remain, reject the request.
	if tokens <= 0 {
		return false, nil
	}

	// Deduct one token and update the bucket.
	tokens--
	err = tb.setBucket(apiKey, tokens, lastRefill)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Middleware applies the token bucket limiter to incoming requests.
func (tb *TokenBucketLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			http.Error(w, "API key required", http.StatusUnauthorized)
			return
		}
		allowed, err := tb.Allow(apiKey)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if !allowed {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
