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
