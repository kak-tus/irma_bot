package storage

import (
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
	"github.com/kak-tus/irma_bot/cnf"
	"go.uber.org/zap"
)

func NewStorage(c *cnf.Cnf, log *zap.SugaredLogger) (*InstanceObj, error) {
	addrs := strings.Split(c.Storage.RedisAddrs, ",")

	var parsed []string
	var pass string

	for _, a := range addrs {
		opt, err := redis.ParseURL(a)

		if err == nil {
			parsed = append(parsed, opt.Addr)
			pass = opt.Password
		}
	}

	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        parsed,
		Password:     pass,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	})

	inst := &InstanceObj{
		cnf: c,
		enc: jsoniter.Config{UseNumber: true}.Froze(),
		log: log,
		rdb: rdb,
	}

	return inst, nil
}

func (o *InstanceObj) Stop() error {
	o.log.Info("Stop storage")

	err := o.rdb.Close()
	if err != nil {
		return err
	}

	o.log.Info("Stopped storage")

	return nil
}
