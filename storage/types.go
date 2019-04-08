package storage

import (
	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

type InstanceObj struct {
	cnf instanceConf
	log *zap.SugaredLogger
	rdb *redis.ClusterClient
}

type instanceConf struct {
	RedisAddrs string
}
