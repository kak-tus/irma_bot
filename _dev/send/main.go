package main

import (
	"context"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kak-tus/irma_bot/config"
	"github.com/kak-tus/irma_bot/model"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
)

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
		log.Panic().Err(err).Msg("fail load model")
	}

	bot, err := tgbotapi.NewBotAPI(cnf.Telegram.Token)
	if err != nil {
		log.Panic().Err(err).Msg("fail load bot")
	}

	ids, err := modelHdl.Queries.GetGroups(context.Background())
	if err != nil {
		log.Panic().Err(err).Msg("fail load bot")
	}

	txt, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Panic().Err(err).Msg("fail load bot")
	}

	for _, id := range ids {
		println(id)

		msg := tgbotapi.NewMessage(id, string(txt))

		resp, err := bot.Send(msg)
		if err != nil {
			log.Error().Err(err).Interface("resp", resp).Msg("send failed")
		}

		time.Sleep(time.Second)
	}
}
