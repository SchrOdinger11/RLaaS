// In redis/client.go
package redis

import (
	"context"
	"fmt"
	"os"
	"sync"

	redispkg "github.com/go-redis/redis/v8"
)

var (
	Ctx    = context.Background()
	client *redispkg.Client
	once   sync.Once
)

type Client = redispkg.Client

// Nil re-exports the external Redis Nil error.
const Nil = redispkg.Nil

// InitClient initializes the Redis client with the given host and port from environment variables.
func InitClient() *redispkg.Client {
	once.Do(func() {
		host := os.Getenv("REDIS_HOST")
		if host == "" {
			host = "127.0.0.1"
		}
		port := os.Getenv("REDIS_PORT")
		if port == "" {
			port = "6379"
		}
		addr := fmt.Sprintf("%s:%s", host, port)
		client = redispkg.NewClient(&redispkg.Options{
			Addr: addr,
		})
	})
	return client
}

// GetClient returns the singleton Redis client.
func GetClient() *redispkg.Client {
	return client
}
