package redis

import (
	"context"
	"sync"

	redispkg "github.com/go-redis/redis/v8"
)

// Client is an alias for the Redis client from github.com/go-redis/redis/v8.
type Client = redispkg.Client

var (
	ctx    = context.Background()
	client *Client
	once   sync.Once
)

// Nil re-exports the external Redis Nil error for use in other packages.
const Nil = redispkg.Nil

// GetClient returns a singleton Redis client instance.
// It creates the client only once, even if called multiple times.
func GetClient(addr string) *Client {
	once.Do(func() {
		client = redispkg.NewClient(&redispkg.Options{
			Addr: addr, // e.g., "localhost:6379"
		})
	})
	return client
}
