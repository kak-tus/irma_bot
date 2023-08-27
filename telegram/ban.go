package telegram

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const nameLimit = 100

func (hdl *InstanceObj) banLongNames(msg *tgbotapi.Message) (bool, error) {
	if msg.NewChatMembers == nil {
		return false, nil
	}

	var toDel bool

	for _, u := range msg.NewChatMembers {
		if len(u.FirstName) >= nameLimit || len(u.LastName) >= nameLimit {
			hdl.oldLog.Infow("Ban long name",
				"User", u.FirstName,
				"Chat", msg.Chat.ID,
			)

			toDel = true

			break
		}
	}

	if !toDel {
		return false, nil
	}

	for _, u := range msg.NewChatMembers {
		kick := tgbotapi.KickChatMemberConfig{
			ChatMemberConfig: tgbotapi.ChatMemberConfig{
				ChatID: msg.Chat.ID,
				UserID: u.ID,
			},
			UntilDate: time.Now().In(time.UTC).AddDate(0, 0, 1).Unix(),
		}

		_, err := hdl.bot.Request(kick)
		if err != nil {
			return true, err
		}
	}

	if err := hdl.deleteMessage(msg.Chat.ID, msg.MessageID); err != nil {
		return true, err
	}

	return true, nil
}

func (hdl *InstanceObj) banKickPool(ctx context.Context, msg *tgbotapi.Message) (bool, error) {
	kicked, err := hdl.stor.IsKicked(ctx, msg.Chat.ID, int(msg.From.ID))
	if err != nil {
		return false, err
	}

	if !kicked {
		return false, err
	}

	hdl.oldLog.Infow("User found in kick pool",
		"User", msg.From.FirstName,
		"Chat", msg.Chat.ID,
	)

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
