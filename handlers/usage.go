package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	redisClient "github.com/SchrOdinger11/RLaaS/redis"
)

// UsageInfo represents the current usage details for a client.
type UsageInfo struct {
	CurrentCount   int   `json:"current_count"`
	Limit          int   `json:"limit"`
	Window         int   `json:"window"`           // in seconds
	TimeUntilReset int64 `json:"time_until_reset"` // in seconds
}

// UsageHandler returns usage information for the client.
func UsageHandler(w http.ResponseWriter, r *http.Request) {
	var Ctx = context.Background()
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "API key required", http.StatusUnauthorized)
		return
	}

	// Retrieve configuration from Redis.
	client := redisClient.GetClient("localhost:6379")
	configKey := "config:" + apiKey
	configMap, err := client.HGetAll(Ctx, configKey).Result()
	if err != nil || len(configMap) == 0 {
		http.Error(w, "No configuration found for API key", http.StatusNotFound)
		return
	}

	limit, err := strconv.Atoi(configMap["limit"])
	if err != nil {
		http.Error(w, "Invalid config limit", http.StatusInternalServerError)
		return
	}

	window, err := strconv.Atoi(configMap["window"])
	if err != nil {
		http.Error(w, "Invalid config window", http.StatusInternalServerError)
		return
	}

	// Determine the current window and retrieve the usage counter.
	now := time.Now().Unix()
	windowStart := now - (now % int64(window))
	counterKey := fmt.Sprintf("rate:%s:%d", apiKey, windowStart)
	currentCount, err := client.Get(Ctx, counterKey).Int()
	if err != nil && err != redisClient.Nil {
		http.Error(w, "Error retrieving usage", http.StatusInternalServerError)
		return
	}

	// Calculate how many seconds until the current window resets.
	elapsed := now - windowStart
	timeUntilReset := int64(window) - elapsed

	usage := UsageInfo{
		CurrentCount:   currentCount,
		Limit:          limit,
		Window:         window,
		TimeUntilReset: timeUntilReset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

// TokenBucketUsageInfo represents the token bucket usage details.
type TokenBucketUsageInfo struct {
	CurrentTokens int   `json:"current_tokens"`
	Capacity      int   `json:"capacity"`
	LastRefill    int64 `json:"last_refill"` // Unix timestamp
	// Optionally, you can include time until full refill
	TimeUntilFull int64 `json:"time_until_full"`
}

// TokenBucketUsageHandler returns the token bucket usage information.
func TokenBucketUsageHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "API key required", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	client := redisClient.GetClient("localhost:6379")
	key := "bucket:" + apiKey

	result, err := client.HGetAll(ctx, key).Result()
	if err != nil {
		http.Error(w, "Error retrieving usage", http.StatusInternalServerError)
		return
	}

	// If no bucket exists, assume full bucket.
	if len(result) == 0 {
		usage := TokenBucketUsageInfo{
			CurrentTokens: 5, // assuming default capacity, change as needed
			Capacity:      5,
			LastRefill:    time.Now().Unix(),
			TimeUntilFull: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(usage)
		return
	}

	tokens, _ := strconv.Atoi(result["tokens"])
	capacity, _ := strconv.Atoi(result["capacity"])
	lastRefill, _ := strconv.ParseInt(result["last_refill"], 10, 64)
	// Calculate time until full refill (this is an approximation)
	refillRate := 1 // assuming 1 token per second; adjust accordingly
	timeUntilFull := int64((capacity - tokens) / refillRate)

	usage := TokenBucketUsageInfo{
		CurrentTokens: tokens,
		Capacity:      capacity,
		LastRefill:    lastRefill,
		TimeUntilFull: timeUntilFull,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}
