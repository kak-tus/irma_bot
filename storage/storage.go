package storage

import (
	"math/rand"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
	"github.com/kak-tus/irma_bot/cnf"
	regen "github.com/zach-klippenstein/goregen"
	"go.uber.org/zap"
)

type InstanceObj struct {
	enc jsoniter.API
	gen regen.Generator
	log *zap.SugaredLogger
	rdb *redis.ClusterClient
}

type Options struct {
	Log    *zap.SugaredLogger
	Config cnf.Stor
}

func NewStorage(opts Options) (*InstanceObj, error) {
	addrs := strings.Split(opts.Config.RedisAddrs, ",")

	var (
		parsed []string
		pass   string
	)

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

	arg := &regen.GeneratorArgs{
		RngSource: rand.NewSource(time.Now().UnixNano()),
	}

	gen, err := regen.NewGenerator("[A-Za-z0-9]{128}", arg)
	if err != nil {
		return nil, err
	}

	inst := &InstanceObj{
		enc: jsoniter.Config{UseNumber: true}.Froze(),
		gen: gen,
		log: opts.Log,
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
