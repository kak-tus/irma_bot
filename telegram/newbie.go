package telegram

import (
	"fmt"
	"math/rand"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kak-tus/irma_bot/storage"
)

func (o *InstanceObj) messageFromNewbie(msg *tgbotapi.Message) error {
	var ban bool

	if msg.Entities != nil {
		for _, e := range *msg.Entities {
			if e.Type == "url" || e.Type == "text_link" || e.Type == "mention" || e.Type == "email" {
				ban = true
				break
			}
		}
	}

	if msg.CaptionEntities != nil {
		for _, e := range *msg.CaptionEntities {
			if e.Type == "url" || e.Type == "text_link" || e.Type == "mention" || e.Type == "email" {
				ban = true
				break
			}
		}
	}

	if msg.ForwardFrom != nil || msg.ForwardFromChat != nil || msg.Sticker != nil || msg.Photo != nil {
		ban = true
	}

	if !ban {
		return o.stor.AddNewbieMessages(msg.Chat.ID, msg.From.ID)
	}

	o.log.Infow("Restricted message",
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

	_, err := o.bot.KickChatMember(kick)
	if err != nil {
		return err
	}

	if err := o.deleteMessage(msg.Chat.ID, msg.MessageID); err != nil {
		return err
	}

	return nil
}

func (o *InstanceObj) newMembers(msg *tgbotapi.Message) error {
	isAdm, err := o.isAdmin(msg.Chat.ID, msg.From.ID)
	if err != nil {
		return err
	}

	if isAdm {
		o.log.Infow(
			"Newbie added by admin, it is normal",
			"Admin", msg.From.ID,
		)
		return nil
	}

	gr, err := o.db.GetGroup(msg.Chat.ID)
	if err != nil {
		return err
	}

	if gr == nil || (gr.BanURL != nil && *gr.BanURL) {
		for _, m := range *msg.NewChatMembers {
			o.log.Infow("Newbie found, add messages",
				"User", m.FirstName,
				"Chat", msg.Chat.ID,
			)

			err := o.stor.AddNewbieMessages(msg.Chat.ID, m.ID)
			if err != nil {
				return err
			}
		}
	}

	if gr == nil || gr.BanQuestion == nil || !*gr.BanQuestion {
		return nil
	}

	quest := o.cnf.Telegram.DefaultQuestions
	greet := o.cnf.Telegram.DefaultGreeting

	if gr != nil && len(gr.Questions) != 0 {
		quest = gr.Questions
	}
	if gr != nil && gr.Greeting != nil && *gr.Greeting != "" {
		greet = *gr.Greeting
	}

	for _, m := range *msg.NewChatMembers {
		o.log.Infow("Newbie found, send question",
			"User", m.FirstName,
			"Chat", msg.Chat.ID,
		)

		qID := rand.Intn(len(quest))

		var name string
		if m.UserName != "" {
			name = m.UserName
		} else {
			name = m.FirstName

			if m.LastName != "" {
				name += " " + m.LastName
			}
		}

		txt := fmt.Sprintf("@%s %s\n\n%s", name, greet, quest[qID].Text)

		resp := tgbotapi.NewMessage(msg.Chat.ID, txt)

		btns := make([][]tgbotapi.InlineKeyboardButton, len(quest[qID].Answers))

		for i, a := range quest[qID].Answers {
			id := fmt.Sprintf("%d_%d_%d_%d", m.ID, msg.Chat.ID, qID, i)
			btns[i] = []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(a.Text, id)}
		}

		resp.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(btns...)

		res, err := o.bot.Send(resp)
		if err != nil {
			return err
		}

		banTimeout := o.cnf.Telegram.DefaultBanTimeoutParsed
		if gr != nil && gr.BanTimeout != nil {
			banTimeout = *gr.BanTimeout
		}

		act := storage.Action{
			ChatID:    res.Chat.ID,
			Type:      "del",
			MessageID: res.MessageID,
			UserID:    m.ID,
		}
		if err := o.stor.AddToActionPool(act, banTimeout); err != nil {
			return err
		}

		err = o.stor.SetKicked(msg.Chat.ID, m.ID, banTimeout)
		if err != nil {
			return err
		}

		act = storage.Action{
			ChatID:    msg.Chat.ID,
			Type:      "del",
			MessageID: msg.MessageID,
			UserID:    m.ID,
		}
		if err := o.stor.AddToActionPool(act, banTimeout); err != nil {
			return err
		}

		act = storage.Action{
			ChatID: msg.Chat.ID,
			Type:   "kick",
			UserID: m.ID,
		}
		if err := o.stor.AddToActionPool(act, banTimeout); err != nil {
			return err
		}
	}

	return nil
}
