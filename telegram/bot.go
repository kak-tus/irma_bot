package telegram

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (o *InstanceObj) messageToBot(msg *tgbotapi.Message) error {
	adms, err := o.bot.GetChatAdministrators(tgbotapi.ChatConfig{ChatID: msg.Chat.ID})
	if err != nil {
		return err
	}

	var foundAdm bool

	for _, a := range adms {
		if a.User != nil && a.User.ID == msg.From.ID {
			foundAdm = true
			break
		}
	}

	if !foundAdm {
		return nil
	}

	for k, v := range o.cnf.Texts.Commands {
		if !strings.Contains(msg.Text, k) {
			continue
		}

		o.log.Debugf("Command %s", k)

		err := o.sett.CreateGroup(msg.Chat.ID, map[string]interface{}{v.Field: v.Value})
		if err != nil {
			return err
		}

		resp := tgbotapi.NewMessage(msg.Chat.ID, v.Text)

		_, err = o.bot.Send(resp)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}
