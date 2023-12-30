package model

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kak-tus/irma_bot/model/queries"
	"github.com/rs/zerolog"
)

type Options struct {
	Log zerolog.Logger
	URL string
}

type Model struct {
	conn    *pgxpool.Pool
	log     zerolog.Logger
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
	hdl.log.Info().Msg("stop model")

	hdl.conn.Close()

	hdl.log.Info().Msg("stopped model")
}
