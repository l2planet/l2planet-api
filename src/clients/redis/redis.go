package redis

import (
	"github.com/go-redis/redis/v9"
)

var client *Client

type Client struct {
	*redis.Client
}

func GetClient() *Client {
	if client == nil {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     "redis:6379",
			Password: "",
			DB:       0,
		})
		client = &Client{
			Client: redisClient,
		}
	}

	return client
}
