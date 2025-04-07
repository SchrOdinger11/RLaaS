package main

import (
	"log"
	"net/http"
	"time"

	handlers "github.com/SchrOdinger11/RLaaS/handlers"
	middleware "github.com/SchrOdinger11/RLaaS/middleware"

	redisClient "github.com/SchrOdinger11/RLaaS/redis"
)

func main() {
	// Initialize the Redis client.
	client := redisClient.InitClient()

	// Set up the rate limiter: for example, 100 requests per minute.
	// limiter := &middleware.RateLimiter{
	// 	RedisClient:   client,
	// 	DefaultLimit:  5,
	// 	DefaultWindow: time.Minute,
	// }
	// Option 2: Use Token Bucket Limiter
	limiter := &middleware.TokenBucketLimiter{
		RedisClient:  client,
		Capacity:     5,           // maximum tokens
		RefillRate:   1,           // 1 token per second
		RefillWindow: time.Second, // refill period
	}

	// Set up HTTP server and endpoints.
	mux := http.NewServeMux()
	checkHandler := http.HandlerFunc(handlers.CheckHandler)
	// Wrap the /check endpoint with our rate limiter middleware.
	mux.Handle("/check", limiter.Middleware(checkHandler))
	mux.HandleFunc("/configure", handlers.ConfigureBucketHandler)
	mux.HandleFunc("/usage", handlers.TokenBucketUsageHandler)
	log.Println("RLaaS service starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
