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
	client := redisClient.GetClient("localhost:6379")

	// Set up the rate limiter: for example, 100 requests per minute.
	limiter := &middleware.RateLimiter{
		RedisClient: client,
		Limit:       5,
		Window:      time.Minute,
	}

	// Set up HTTP server and endpoints.
	mux := http.NewServeMux()
	checkHandler := http.HandlerFunc(handlers.CheckHandler)
	// Wrap the /check endpoint with our rate limiter middleware.
	mux.Handle("/check", limiter.Middleware(checkHandler))

	log.Println("RLaaS service starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
