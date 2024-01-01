package model

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestGetGroup(t *testing.T) {
	hdl, err := NewTestModel()
	require.NoError(t, err)

	defer hdl.Stop()

	group, err := hdl.Queries.GetGroup(context.Background(), -1001175783418)
	require.NoError(t, err)
	require.True(t, group.BanQuestion.Bool)

	_, err = hdl.Queries.GetGroup(context.Background(), 1)
	require.Error(t, err)
	require.True(t, errors.Is(err, pgx.ErrNoRows))
	require.False(t, errors.Is(err, sql.ErrNoRows))
}

func NewTestModel() (*Model, error) {
	addr := os.Getenv("IRMA_DB_ADDR")
	if addr == "" {
		return nil, nil
	}

	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	modelOpts := Options{
		Log: log,
		URL: addr,
	}

	modelHdl, err := New(modelOpts)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	if err := modelHdl.Start(ctx); err != nil {
		log.Panic().Err(err).Msg("fail start model")
	}

	return modelHdl, nil
}
