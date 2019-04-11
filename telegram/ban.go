package telegram

import (
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (o *InstanceObj) banLongNames(msg *tgbotapi.Message) (bool, error) {
	if msg.NewChatMembers == nil {
		return false, nil
	}

	var toDel bool

	for _, u := range *msg.NewChatMembers {
		if len(u.FirstName) >= o.cnf.NameLimit || len(u.LastName) >= o.cnf.NameLimit {
			o.log.Infof("Ban long name: %s", u.FirstName)
			toDel = true
			break
		}
	}

	if !toDel {
		return false, nil
	}

	for _, u := range *msg.NewChatMembers {
		kick := tgbotapi.KickChatMemberConfig{
			ChatMemberConfig: tgbotapi.ChatMemberConfig{
				ChatID: msg.Chat.ID,
				UserID: u.ID,
			},
			UntilDate: time.Now().In(time.UTC).AddDate(0, 0, 1).Unix(),
		}

		_, err := o.bot.KickChatMember(kick)
		if err != nil {
			return true, err
		}
	}

	del := tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID)

	_, err := o.bot.DeleteMessage(del)
	if err != nil {
		return true, err
	}

	return true, nil
}

func (o *InstanceObj) banKickPool(msg *tgbotapi.Message) (bool, error) {
	kicked, err := o.stor.IsKicked(msg.Chat.ID, msg.From.ID)
	if err != nil {
		return false, err
	}

	if !kicked {
		return false, err
	}

	o.log.Infof("User found in kick pool: %s", msg.From.FirstName)

	kick := tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: msg.Chat.ID,
			UserID: msg.From.ID,
		},
		UntilDate: time.Now().In(time.UTC).AddDate(0, 0, 1).Unix(),
	}

	_, err = o.bot.KickChatMember(kick)
	if err != nil {
		return true, err
	}

	return true, nil
}
