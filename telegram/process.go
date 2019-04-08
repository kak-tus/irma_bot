package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (o *InstanceObj) process(tgbotapi.Update) error {
	println(1)
	return nil
}
