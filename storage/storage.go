package storage

import (
	"strings"
	"time"

	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/applog"
	"git.aqq.me/go/app/event"
	"github.com/go-redis/redis"
	"github.com/iph0/conf"
	jsoniter "github.com/json-iterator/go"
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
				enc: jsoniter.Config{UseNumber: true}.Froze(),
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

			err := inst.rdb.Close()
			if err != nil {
				return err
			}

			inst.log.Info("Stopped storage")
			return nil
		},
	)
}

func Get() *InstanceObj {
	return inst
}
