package db

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
	"github.com/kak-tus/irma_bot/cnf"
	"go.uber.org/zap"
)

type InstanceObj struct {
	cnf  *cnf.Cnf
	log  *zap.SugaredLogger
	pool *pgx.ConnPool
	sqrl sq.StatementBuilderType
}
