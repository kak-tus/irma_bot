package telegram

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kak-tus/irma_bot/storage"
)

func (hdl *InstanceObj) messageFromNewbie(ctx context.Context, msg *tgbotapi.Message) error {
	var ban bool

	if msg.Entities != nil {
		for _, e := range msg.Entities {
			if e.Type == "url" || e.Type == "text_link" || e.Type == "mention" || e.Type == "email" {
				ban = true
				break
			}
		}
	}

	if msg.CaptionEntities != nil {
		for _, e := range msg.CaptionEntities {
			if e.Type == "url" || e.Type == "text_link" || e.Type == "mention" || e.Type == "email" {
				ban = true
				break
			}
		}
	}

	if msg.ForwardFrom != nil ||
		msg.ForwardFromChat != nil ||
		msg.Sticker != nil ||
		msg.Photo != nil ||
		msg.Animation != nil ||
		msg.Audio != nil ||
		msg.Video != nil ||
		msg.VideoNote != nil ||
		msg.Voice != nil {
		ban = true
	}

	if !ban {
		return hdl.stor.AddNewbieMessages(ctx, msg.Chat.ID, int(msg.From.ID))
	}

	hdl.log.Infow("Restricted message",
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

	_, err := hdl.bot.Request(kick)
	if err != nil {
		return err
	}

	if err := hdl.deleteMessage(msg.Chat.ID, msg.MessageID); err != nil {
		return err
	}

	return nil
}

func (hdl *InstanceObj) newMembers(ctx context.Context, msg *tgbotapi.Message) error {
	isAdm, err := hdl.isAdmin(msg.Chat.ID, msg.From.ID)
	if err != nil {
		return err
	}

	if isAdm {
		hdl.log.Infow(
			"Newbie added by admin, it is normal",
			"Admin", msg.From.ID,
		)

		return nil
	}

	gr, err := hdl.model.Queries.GetGroup(ctx, msg.Chat.ID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	defaultGroup := hdl.model.GetDefaultGroup()

	for _, m := range msg.NewChatMembers {
		hdl.log.Infow("Newbie found, add messages",
			"User", m.FirstName,
			"Chat", msg.Chat.ID,
		)

		err := hdl.stor.AddNewbieMessages(ctx, msg.Chat.ID, int(m.ID))
		if err != nil {
			return err
		}
	}

	// Ban by question by default if group is not registered
	if gr.BanQuestion.Valid && !gr.BanQuestion.Bool {
		return nil
	}

	quest := defaultGroup.Questions.Questions
	greet := defaultGroup.Greeting.String

	if len(gr.Questions.Questions) != 0 {
		quest = gr.Questions.Questions
	}

	if gr.Greeting.Valid {
		greet = gr.Greeting.String
	}

	for _, newMember := range msg.NewChatMembers {
		hdl.log.Infow("Newbie found, send question",
			"User", newMember.FirstName,
			"Chat", msg.Chat.ID,
		)

		qID := rand.Intn(len(quest))

		var name string
		if newMember.UserName != "" {
			name = newMember.UserName
		} else {
			name = newMember.FirstName

			if newMember.LastName != "" {
				name += " " + newMember.LastName
			}
		}

		txt := fmt.Sprintf("@%s %s\n\n%s", name, greet, quest[qID].Text)

		resp := tgbotapi.NewMessage(msg.Chat.ID, txt)

		btns := make([][]tgbotapi.InlineKeyboardButton, len(quest[qID].Answers))

		for i, a := range quest[qID].Answers {
			id := fmt.Sprintf("%d_%d_%d_%d", newMember.ID, msg.Chat.ID, qID, i)
			btns[i] = []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(a.Text, id)}
		}

		resp.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(btns...)

		res, err := hdl.bot.Send(resp)
		if err != nil {
			return err
		}

		banTimeout := time.Duration(defaultGroup.BanTimeout.Int32) * time.Minute
		if gr.BanTimeout.Valid {
			banTimeout = time.Duration(gr.BanTimeout.Int32) * time.Minute
		}

		act := storage.Action{
			ChatID:    res.Chat.ID,
			Type:      storage.ActionTypeDelete,
			MessageID: res.MessageID,
			UserID:    int(newMember.ID),
		}
		if err := hdl.stor.AddToActionPool(ctx, act, banTimeout); err != nil {
			return err
		}

		err = hdl.stor.SetKicked(ctx, msg.Chat.ID, int(newMember.ID), banTimeout)
		if err != nil {
			return err
		}

		act = storage.Action{
			ChatID:    msg.Chat.ID,
			Type:      storage.ActionTypeDelete,
			MessageID: msg.MessageID,
			UserID:    int(newMember.ID),
		}
		if err := hdl.stor.AddToActionPool(ctx, act, banTimeout); err != nil {
			return err
		}

		act = storage.Action{
			ChatID: msg.Chat.ID,
			Type:   storage.ActionTypeKick,
			UserID: int(newMember.ID),
		}
		if err := hdl.stor.AddToActionPool(ctx, act, banTimeout); err != nil {
			return err
		}
	}

	return nil
}
