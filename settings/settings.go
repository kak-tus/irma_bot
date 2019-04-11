package settings

import (
	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/applog"
	"git.aqq.me/go/app/event"
	sq "github.com/Masterminds/squirrel"
	"github.com/iph0/conf"
	"github.com/jackc/pgx"
)

var inst *InstanceObj

func init() {
	event.Init.AddHandler(
		func() error {
			cnfMap := appconf.GetConfig()["settings"]

			var cnf instanceConf
			err := conf.Decode(cnfMap, &cnf)
			if err != nil {
				return err
			}

			connCnf, err := pgx.ParseConnectionString(cnf.DBAddr)
			if err != nil {
				return err
			}

			pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
				ConnConfig:     connCnf,
				MaxConnections: 10,
			})
			if err != nil {
				return err
			}

			sqrl := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

			inst = &InstanceObj{
				cnf:  cnf,
				log:  applog.GetLogger().Sugar(),
				pool: pool,
				sqrl: sqrl,
			}

			inst.log.Info("Started settings")

			return nil
		},
	)

	event.Stop.AddHandler(
		func() error {
			inst.log.Info("Stop settings")

			inst.pool.Close()

			inst.log.Info("Stopped settings")
			return nil
		},
	)
}

func Get() *InstanceObj {
	return inst
}
