package redis

import (
	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/doug-martin/goqu/v9"
	"go.uber.org/zap"
)

type RedisPubSub struct {
	PubSubModel *PubSubModel
}

func InitRedisPubSub(db *goqu.Database, pubSubCfg config.RedisClientConfig, logger *zap.Logger) (*RedisPubSub, error) {

	pubSubClientModel, err := InitPubSubModel(pubSubCfg.RedisAddr+":"+pubSubCfg.RedisPort, pubSubCfg.RedisPass, pubSubCfg.RedisDb)
	if err != nil {
		return nil, err
	}

	return &RedisPubSub{
		PubSubModel: pubSubClientModel,
	}, nil
}
