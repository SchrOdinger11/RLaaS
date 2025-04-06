package handlers

import (
	"context"
	"encoding/json"
	"net/http"

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
	client := redisClient.GetClient("localhost:6379")
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
