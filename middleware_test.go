package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	handlers "github.com/SchrOdinger11/RLaaS/handlers"
	middleware "github.com/SchrOdinger11/RLaaS/middleware"
	redisClient "github.com/SchrOdinger11/RLaaS/redis"
)

func TestRateLimiterMiddleware(t *testing.T) {
	// Initialize a test Redis client (or use a mock Redis).
	client := redisClient.GetClient("localhost:6379")
	limiter := &middleware.RateLimiter{
		RedisClient:   client,
		DefaultLimit:  2, // set a low limit for testing
		DefaultWindow: time.Minute,
	}

	// Create a request with an API key.
	req, err := http.NewRequest("GET", "/check", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-API-Key", "test-key")

	// Use ResponseRecorder to capture the output.
	rr := httptest.NewRecorder()

	// Wrap the dummyHandler with our middleware.
	handler := limiter.Middleware(http.HandlerFunc(handlers.CheckHandler))

	// First request should be allowed.
	handler.ServeHTTP(rr, req)
	if strings.TrimSpace(rr.Body.String()) != "Request Allowed" {
		t.Errorf("expected OK, got %s", rr.Body.String())
	}

	// Second request.
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if strings.TrimSpace(rr.Body.String()) != "Request Allowed" {
		t.Errorf("expected OK, got %s", rr.Body.String())
	}

	// Third request should exceed the limit.
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code == http.StatusTooManyRequests {
		t.Errorf("expected status 429 Too Many Requests, got %d", rr.Code)
	}
}
