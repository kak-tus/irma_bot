package db

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
	"github.com/kak-tus/irma_bot/cnf"
	"go.uber.org/zap"
)

func NewDB(c *cnf.Cnf, log *zap.SugaredLogger) (*InstanceObj, error) {
	connCnf, err := pgx.ParseConnectionString(c.DB.DBAddr)
	if err != nil {
		return nil, err
	}

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     connCnf,
		MaxConnections: 10,
	})
	if err != nil {
		return nil, err
	}

	sqrl := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	inst := &InstanceObj{
		cnf:  c,
		log:  log,
		pool: pool,
		sqrl: sqrl,
	}
	return inst, nil
}

func (o *InstanceObj) Stop() {
	o.log.Info("Stop DB")

	o.pool.Close()

	o.log.Info("Stopped DB")
}
