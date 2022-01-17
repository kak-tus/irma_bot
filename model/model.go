package model

import (
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib" // pgx
	"github.com/kak-tus/irma_bot/model/queries"
	"go.uber.org/zap"
)

type Options struct {
	Log *zap.SugaredLogger
	URL string
}

type Model struct {
	log     *zap.SugaredLogger
	Queries *queries.Queries
}

func NewModel(opts Options) (*Model, error) {
	conn, err := sql.Open("pgx", opts.URL)
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(2)
	conn.SetMaxIdleConns(2)

	qry := queries.New(conn)

	mdl := &Model{
		log:     opts.Log,
		Queries: qry,
	}

	return mdl, nil
}
