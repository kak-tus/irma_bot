package storage

import (
	"github.com/go-redis/redis"
	jsoniter "github.com/json-iterator/go"
	"github.com/kak-tus/irma_bot/cnf"
	"go.uber.org/zap"
)

type InstanceObj struct {
	cnf *cnf.Cnf
	enc jsoniter.API
	log *zap.SugaredLogger
	rdb *redis.ClusterClient
}
