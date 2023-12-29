package model

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kak-tus/irma_bot/model/queries"
	"go.uber.org/zap"
)

type Options struct {
	Log *zap.SugaredLogger
	URL string
}

type Model struct {
	conn    *pgxpool.Pool
	log     *zap.SugaredLogger
	Queries *queries.Queries
	url     string
}

func New(opts Options) (*Model, error) {
	mdl := &Model{
		log: opts.Log,
		url: opts.URL,
	}

	return mdl, nil
}

func (hdl *Model) Start(ctx context.Context) error {
	conn, err := pgxpool.New(ctx, hdl.url)
	if err != nil {
		return err
	}

	hdl.conn = conn
	hdl.Queries = queries.New(conn)

	return nil
}

func (hdl *Model) Stop() {
	hdl.log.Info("stop model")

	hdl.conn.Close()

	hdl.log.Info("stopped model")
}
