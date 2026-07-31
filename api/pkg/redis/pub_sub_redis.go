package redis

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type PubSubModel struct {
	Ctx    context.Context
	Client redis.Client
}

func InitPubSubModel(addr, password string, db int) (*PubSubModel, error) {
	client := redis.NewClient(&redis.Options{
		Addr:             addr,
		Password:         password,
		DB:               db,
		DisableIndentity: true,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	return &PubSubModel{
		Ctx:    context.Background(),
		Client: *client,
	}, nil
}
