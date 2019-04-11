package telegram

import (
	"fmt"
	"strings"

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

	// Ban users with extra long names
	// It's probably "name spammers"
	banned, err := o.banLongNames(msg)
	if err != nil {
		return err
	}
	if banned {
		return nil
	}

	banned, err = o.banKickPool(msg)
	if err != nil {
		return err
	}
	if banned {
		return nil
	}

	cnt, err := o.stor.GetNewbieMessages(msg.Chat.ID, msg.From.ID)
	if err != nil {
		return err
	}

	// In case of newbie we got count >0, for ordinary user count=0
	if cnt > 0 && cnt <= 4 {
		// if cnt > 0 && cnt <= 40 {
		return o.messageFromNewbie(msg)
	}

	if msg.NewChatMembers != nil {
		return o.newMembers(msg)
	}

	name := fmt.Sprintf("@%s", o.cnf.BotName)

	if strings.HasPrefix(msg.Text, name) {
		return o.messageToBot(msg)
	}

	return nil
}

func (o *InstanceObj) processCallback(msg *tgbotapi.CallbackQuery) error {
	return nil
}
