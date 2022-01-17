package telegram

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kak-tus/irma_bot/model/queries"
)

func (o *InstanceObj) processCommands(ctx context.Context, msg *tgbotapi.Message) (bool, error) {
	params := queries.CreateOrUpdateGroupParametersParams{
		ID: msg.Chat.ID,
	}

	var respMsg string

	switch {
	case strings.Contains(msg.Text, "use_ban_url"):
		params.BanUrl = sql.NullBool{Valid: true, Bool: true}
		respMsg = `URLs protection enabled\nSend me "no_ban_url" to disable`
	case strings.Contains(msg.Text, "no_ban_url"):
		params.BanUrl = sql.NullBool{Valid: true, Bool: false}
		respMsg = `URLs protection disabled\nSend me "use_ban_url" to enable`
	case strings.Contains(msg.Text, "use_ban_question"):
		params.BanUrl = sql.NullBool{Valid: true, Bool: true}
		respMsg = `Questions protection enabled\nSend me "no_ban_question" to disable`
	case strings.Contains(msg.Text, "no_ban_question"):
		params.BanUrl = sql.NullBool{Valid: true, Bool: false}
		respMsg = `Questions protection disabled\nSend me "use_ban_question" to enable`
	case strings.Contains(msg.Text, "ban_timeout"):
		idx := strings.Index(msg.Text, "ban_timeout")

		// already checked on contains
		toParse := strings.TrimSpace(msg.Text[idx+len("ban_timeout"):])

		parsed, err := strconv.Atoi(toParse)
		if err != nil {
			return false, err
		}

		if parsed < 1 || parsed > 60 {
			return false, nil
		}

		params.BanTimeout = sql.NullInt32{Valid: true, Int32: int32(parsed)}
		respMsg = "Ban timeout setuped"
	default:
		return false, nil
	}

	err := o.model.Queries.CreateOrUpdateGroupParameters(ctx, params)
	if err != nil {
		return false, err
	}

	resp := tgbotapi.NewMessage(msg.Chat.ID, respMsg)

	_, err = o.bot.Send(resp)
	if err != nil {
		return false, err
	}

	return true, nil
}
