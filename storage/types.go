package storage

import (
	"github.com/go-redis/redis"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type InstanceObj struct {
	cnf instanceConf
	enc jsoniter.API
	log *zap.SugaredLogger
	rdb *redis.ClusterClient
}

type instanceConf struct {
	RedisAddrs string
}
