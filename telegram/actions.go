package telegram

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kak-tus/irma_bot/storage"
)

func (hdl *InstanceObj) processActions() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	acts, err := hdl.stor.GetFromActionPool(ctx)
	if err != nil {
		return err
	}

	for _, a := range acts {
		kicked, err := hdl.stor.IsKicked(ctx, a.ChatID, a.UserID)
		if err != nil {
			return err
		}

		if !kicked {
			continue
		}

		if a.Type == storage.ActionTypeDelete {
			if err := hdl.deleteMessage(a.ChatID, a.MessageID); err != nil {
				return err
			}
		} else if a.Type == storage.ActionTypeKick {
			kick := tgbotapi.KickChatMemberConfig{
				ChatMemberConfig: tgbotapi.ChatMemberConfig{
					ChatID: a.ChatID,
					UserID: int64(a.UserID),
				},
				UntilDate: time.Now().In(time.UTC).AddDate(0, 0, 1).Unix(),
			}

			_, err := hdl.bot.Request(kick)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
