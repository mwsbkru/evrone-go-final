package redis

import (
	"github.com/mwsbkru/evrone-go-final/config"

	"github.com/redis/go-redis/v9"
)

// Client wraps Redis client functionality
type Client struct {
	client *redis.Client
}

// NewClient creates a new Redis client based on the provided configuration
func NewClient(cfg *config.Config) *Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
		DB:   cfg.Redis.DB,
	})

	return &Client{client: redisClient}
}

// Close closes the Redis client
func (c *Client) Close() error {
	return c.client.Close()
}

// GetClient returns the underlying redis.Client
func (c *Client) GetClient() *redis.Client {
	return c.client
}

