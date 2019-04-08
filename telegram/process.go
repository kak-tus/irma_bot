package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (o *InstanceObj) process(msg tgbotapi.Update) error {
	if msg.Message != nil {
		return o.processMsg(msg.Message)
	} else if msg.CallbackQuery != nil {
		return o.processCallback(msg.CallbackQuery)
	}

	return nil
}

func (o *InstanceObj) processMsg(msg *tgbotapi.Message) error {
	if msg.Chat.IsPrivate() {
		resp := tgbotapi.NewMessage(msg.Chat.ID, o.cnf.Texts.Usage)
		_, err := o.bot.Send(resp)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (o *InstanceObj) processCallback(msg *tgbotapi.CallbackQuery) error {
	return nil
}
