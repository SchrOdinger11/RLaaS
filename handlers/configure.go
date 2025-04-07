package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	redisClient "github.com/SchrOdinger11/RLaaS/redis"
)

// Config represents the rate limiting configuration.
type Config struct {
	Limit  int `json:"limit"`  // Maximum requests allowed
	Window int `json:"window"` // Time window in seconds
}

// ConfigureHandler allows clients to set their rate limiting configuration.
func ConfigureHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "API key required", http.StatusUnauthorized)
		return
	}

	var config Config
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Store configuration in Redis using a hash key "config:<apiKey>"
	client := redisClient.InitClient()
	key := "config:" + apiKey
	_, err := client.HSet(ctx, key, map[string]interface{}{
		"limit":  config.Limit,
		"window": config.Window,
	}).Result()
	if err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Configuration updated"))
}

// BucketConfig represents the configuration for the token bucket.
type BucketConfig struct {
	Capacity     int `json:"capacity"`      // Maximum tokens available.
	RefillRate   int `json:"refill_rate"`   // Tokens added per second.
	RefillWindow int `json:"refill_window"` // Refill interval in seconds (typically 1).
}

// ConfigureBucketHandler allows clients to set their token bucket configuration.
func ConfigureBucketHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "API key required", http.StatusUnauthorized)
		return
	}

	var config BucketConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	Ctx := context.Background()

	// Store configuration in Redis under the key "bucket:<apiKey>"
	client := redisClient.InitClient()
	key := "bucket:" + apiKey
	_, err := client.HSet(Ctx, key, map[string]interface{}{
		"tokens":        config.Capacity, // initialize with full capacity
		"capacity":      config.Capacity,
		"last_refill":   time.Now().Unix(),
		"refill_rate":   config.RefillRate,
		"refill_window": config.RefillWindow,
	}).Result()
	if err != nil {
		http.Error(w, "Failed to save bucket config", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Bucket configuration updated"))
}
