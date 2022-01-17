package telegram

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (o *InstanceObj) processActions() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	acts, err := o.stor.GetFromActionPool(ctx)
	if err != nil {
		return err
	}

	for _, a := range acts {
		kicked, err := o.stor.IsKicked(ctx, a.ChatID, a.UserID)
		if err != nil {
			return err
		}

		if !kicked {
			continue
		}

		if a.Type == "del" {
			if err := o.deleteMessage(a.ChatID, a.MessageID); err != nil {
				return err
			}
		} else if a.Type == "kick" {
			kick := tgbotapi.KickChatMemberConfig{
				ChatMemberConfig: tgbotapi.ChatMemberConfig{
					ChatID: a.ChatID,
					UserID: int64(a.UserID),
				},
				UntilDate: time.Now().In(time.UTC).AddDate(0, 0, 1).Unix(),
			}

			_, err := o.bot.Request(kick)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
