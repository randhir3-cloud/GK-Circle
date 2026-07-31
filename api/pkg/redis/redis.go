package redis

import (
	"time"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/doug-martin/goqu/v9"
	"go.uber.org/zap"
)

type RedisPubSub struct {
	PubSubModel *PubSubModel
}

func InitRedisPubSub(db *goqu.Database, pubSubCfg config.RedisClientConfig, logger *zap.Logger) (*RedisPubSub, error) {
	start := time.Now()
	logger.Info("Redis connection start",
		zap.String("host", pubSubCfg.RedisAddr),
		zap.String("port", pubSubCfg.RedisPort),
		zap.Int("db", pubSubCfg.RedisDb),
	)

	pubSubClientModel, err := InitPubSubModel(pubSubCfg.RedisAddr+":"+pubSubCfg.RedisPort, pubSubCfg.RedisPass, pubSubCfg.RedisDb)
	elapsed := time.Since(start)

	if err != nil {
		logger.Error("Redis connection failure",
			zap.String("host", pubSubCfg.RedisAddr),
			zap.String("port", pubSubCfg.RedisPort),
			zap.Int("db", pubSubCfg.RedisDb),
			zap.Duration("elapsed", elapsed),
			zap.Error(err),
		)
		return nil, err
	}

	logger.Info("Redis connection success",
		zap.String("host", pubSubCfg.RedisAddr),
		zap.String("port", pubSubCfg.RedisPort),
		zap.Int("db", pubSubCfg.RedisDb),
		zap.Duration("elapsed", elapsed),
	)

	return &RedisPubSub{
		PubSubModel: pubSubClientModel,
	}, nil
}
