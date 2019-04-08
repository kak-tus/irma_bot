package storage

import (
	"fmt"
	"strings"
	"time"

	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/applog"
	"git.aqq.me/go/app/event"
	"github.com/go-redis/redis"
	"github.com/iph0/conf"
)

var inst *InstanceObj

func init() {
	event.Init.AddHandler(
		func() error {
			cnfMap := appconf.GetConfig()["storage"]

			var cnf instanceConf
			err := conf.Decode(cnfMap, &cnf)
			if err != nil {
				return err
			}

			addrs := strings.Split(cnf.RedisAddrs, ",")

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

			inst = &InstanceObj{
				cnf: cnf,
				log: applog.GetLogger().Sugar(),
				rdb: rdb,
			}

			inst.log.Info("Started storage")

			return nil
		},
	)

	event.Stop.AddHandler(
		func() error {
			inst.log.Info("Stop storage")

			inst.log.Info("Stopped storage")
			return nil
		},
	)
}

func Get() *InstanceObj {
	return inst
}

func (o *InstanceObj) IsKicked(chatID int64, userID int) (bool, error) {
	key := fmt.Sprintf("irma_kick_{%d_%d}", chatID, userID)

	res, err := o.rdb.Exists(key).Result()
	if err != nil {
		return false, err
	}

	if res == 0 {
		return false, nil
	}

	return true, nil
}

func (o *InstanceObj) SetKicked(chatID int64, userID int) error {
	key := fmt.Sprintf("irma_kick_{%d_%d}", chatID, userID)

	_, err := o.rdb.Set(key, 1, time.Minute*10).Result()
	if err != nil {
		return err
	}

	return nil
}
