package model

import (
	"context"
	"os"
	"testing"

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
