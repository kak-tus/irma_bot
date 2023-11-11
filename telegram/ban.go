package telegram

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
)

const nameLimit = 100

func (hdl *InstanceObj) banLongNames(log zerolog.Logger, chatID int64, users []tgbotapi.User) (bool, error) {
	if len(users) == 0 {
		return false, nil
	}

	var toDel bool

	for _, usr := range users {
		if len(usr.FirstName) >= nameLimit || len(usr.LastName) >= nameLimit {
			log.Info().Str("first_name", usr.FirstName).Str("last_name", usr.LastName).Msg("ban long name")

			toDel = true

			break
		}
	}

	if !toDel {
		return false, nil
	}

	for _, usr := range users {
		kick := tgbotapi.KickChatMemberConfig{
			ChatMemberConfig: tgbotapi.ChatMemberConfig{
				ChatID: chatID,
				UserID: usr.ID,
			},
			UntilDate: time.Now().In(time.UTC).AddDate(0, 0, 1).Unix(),
		}

		_, err := hdl.bot.Request(kick)
		if err != nil {
			return true, err
		}
	}

	return true, nil
}

func (hdl *InstanceObj) banKickPool(ctx context.Context, log zerolog.Logger, msg *tgbotapi.Message) (bool, error) {
	kicked, err := hdl.stor.IsKicked(ctx, msg.Chat.ID, int(msg.From.ID))
	if err != nil {
		return false, err
	}

	if !kicked {
		return false, err
	}

	log.Info().Str("user", msg.From.FirstName).Msg("user found in kick pool")

	kick := tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: msg.Chat.ID,
			UserID: msg.From.ID,
		},
		UntilDate: time.Now().In(time.UTC).AddDate(0, 0, 1).Unix(),
	}

	_, err = hdl.bot.Request(kick)
	if err != nil {
		return true, err
	}

	if err := hdl.deleteMessage(msg.Chat.ID, msg.MessageID); err != nil {
		return true, err
	}

	return true, nil
}
