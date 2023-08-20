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
	"go.uber.org/zap"
)

//go:generate sqlc generate
//go:generate oapi-codegen --config openapi-codegen.yml openapi.yml

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	log := logger.Sugar()

	cnf, err := config.NewConf()
	if err != nil {
		log.Panic(err)
	}

	modelOpts := model.Options{
		Log: log,
		URL: cnf.DB.Addr,
	}

	modelHdl, err := model.NewModel(modelOpts)
	if err != nil {
		log.Panic(err)
	}

	storOptions := storage.Options{
		Log:    log,
		Config: cnf.Storage,
	}

	stor, err := storage.NewStorage(storOptions)
	if err != nil {
		log.Panic(err)
	}

	apiOpts := api.Options{
		Log:     log,
		Model:   modelHdl,
		Storage: stor,
	}

	apiHdl, err := api.NewAPI(apiOpts)
	if err != nil {
		log.Panic(err)
	}

	telegramOpts := telegram.Options{
		Log:     log,
		Config:  cnf.Telegram,
		Model:   modelHdl,
		Router:  apiHdl.GetHTTPRouter(),
		Storage: stor,
	}

	tg, err := telegram.NewTelegram(telegramOpts)
	if err != nil {
		log.Panic(err)
	}

	go func() {
		if err := tg.Start(); err != nil {
			log.Panic(err)
		}
	}()

	srv := &http.Server{
		Addr:    cnf.Listen,
		Handler: apiHdl.GetHTTPRouter(),
	}

	go func() {
		err = srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Panic(err)
		}
	}()

	st := make(chan os.Signal, 1)
	signal.Notify(st, os.Interrupt)

	<-st
	log.Info("stop")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	err = srv.Shutdown(ctx)
	if err != nil {
		log.Panic(err)
	}

	cancel()

	if err := tg.Stop(); err != nil {
		log.Panic(err)
	}

	_ = log.Sync()
}
