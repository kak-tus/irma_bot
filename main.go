package main

import (
	"os"
	"os/signal"

	"github.com/kak-tus/irma_bot/telegram"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	log := logger.Sugar()

	tg, err := telegram.NewTelegram(log)
	if err != nil {
		log.Panic(err)
	}

	go func() {
		if err := tg.Start(); err != nil {
			log.Panic(err)
		}
	}()

	st := make(chan os.Signal, 1)
	signal.Notify(st, os.Interrupt)

	<-st
	log.Info("Stop")

	if err := tg.Stop(); err != nil {
		log.Panic(err)
	}

	_ = log.Sync()
}
