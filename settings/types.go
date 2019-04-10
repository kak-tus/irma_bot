package settings

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
	"go.uber.org/zap"
)

type InstanceObj struct {
	cnf  instanceConf
	log  *zap.SugaredLogger
	pool *pgx.ConnPool
	sqrl sq.StatementBuilderType
}

type instanceConf struct {
	DBAddr string
}
