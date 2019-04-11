package telegram

import (
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (o *InstanceObj) processActions() error {
	acts, err := o.stor.GetFromActionPool()
	if err != nil {
		return err
	}

	for _, a := range acts {
		kicked, err := o.stor.IsKicked(a.ChatID, a.UserID)
		if err != nil {
			return err
		}

		if !kicked {
			continue
		}

		if a.Type == "del" {
			msg := tgbotapi.NewDeleteMessage(a.ChatID, a.MessageID)
			_, err = o.bot.Send(msg)
			if err != nil {
				return err
			}
		} else if a.Type == "kick" {
			kick := tgbotapi.KickChatMemberConfig{
				ChatMemberConfig: tgbotapi.ChatMemberConfig{
					ChatID: a.ChatID,
					UserID: a.UserID,
				},
				UntilDate: time.Now().In(time.UTC).AddDate(0, 0, 1).Unix(),
			}

			_, err := o.bot.KickChatMember(kick)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
