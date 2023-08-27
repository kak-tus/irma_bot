package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/kak-tus/irma_bot/api"
	"github.com/kak-tus/irma_bot/config"
	"github.com/kak-tus/irma_bot/model"
	"github.com/kak-tus/irma_bot/storage"
	"github.com/kak-tus/irma_bot/telegram"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
)

//go:generate sqlc generate
//go:generate oapi-codegen --config openapi-codegen.yml openapi.yml

func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	oldLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	oldLog := oldLogger.Sugar()

	cnf, err := config.NewConf()
	if err != nil {
		log.Panic().Err(err).Msg("fail load config")
	}

	modelOpts := model.Options{
		Log: oldLog,
		URL: cnf.DB.Addr,
	}

	modelHdl, err := model.NewModel(modelOpts)
	if err != nil {
		oldLog.Panic(err)
	}

	storOptions := storage.Options{
		Log:    oldLog,
		Config: cnf.Storage,
	}

	stor, err := storage.NewStorage(storOptions)
	if err != nil {
		oldLog.Panic(err)
	}

	apiOpts := api.Options{
		Log:     oldLog,
		Model:   modelHdl,
		Storage: stor,
	}

	apiHdl, err := api.NewAPI(apiOpts)
	if err != nil {
		oldLog.Panic(err)
	}

	telegramOpts := telegram.Options{
		OldLog:  oldLog,
		Config:  cnf.Telegram,
		Model:   modelHdl,
		Router:  apiHdl.GetHTTPRouter(),
		Storage: stor,
		Log:     log.With().Str("module", "telegram").Logger(),
	}

	tg, err := telegram.NewTelegram(telegramOpts)
	if err != nil {
		oldLog.Panic(err)
	}

	go func() {
		if err := tg.Start(); err != nil {
			oldLog.Panic(err)
		}
	}()

	srv := &http.Server{
		Addr:    cnf.Listen,
		Handler: apiHdl.GetHTTPRouter(),
	}

	go func() {
		err = srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			oldLog.Panic(err)
		}
	}()

	st := make(chan os.Signal, 1)
	signal.Notify(st, os.Interrupt)

	<-st
	oldLog.Info("stop")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	err = srv.Shutdown(ctx)
	if err != nil {
		oldLog.Panic(err)
	}

	cancel()

	if err := tg.Stop(); err != nil {
		oldLog.Panic(err)
	}

	_ = oldLog.Sync()
}
