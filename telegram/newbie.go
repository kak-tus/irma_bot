package telegram

import (
	"fmt"
	"math/rand"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jackc/pgx/pgtype"
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

	if msg.ForwardFrom != nil || msg.ForwardFromChat != nil || msg.Sticker != nil || msg.Photo != nil {
		ban = true
	}

	if !ban {
		return o.stor.AddNewbieMessages(msg.Chat.ID, msg.From.ID)
	}

	o.log.Infof("Restricted message from user: %s", msg.From.FirstName)

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
	gr, err := o.sett.GetGroup(msg.Chat.ID)
	if err != nil {
		return err
	}

	if gr == nil || gr.BanURL.Bool {
		for _, m := range *msg.NewChatMembers {
			o.log.Infof("Newbie found, add messages: %s", m.FirstName)
			err := o.stor.AddNewbieMessages(msg.Chat.ID, m.ID)
			if err != nil {
				return err
			}
		}
	}

	if gr != nil && gr.BanQuestion.Status == pgtype.Present && gr.BanQuestion.Bool == false {
		return nil
	}

	quest := o.cnf.DefaultQuestions
	greet := o.cnf.DefaultGreeting

	if gr != nil && len(gr.Questions) != 0 {
		quest = gr.Questions
	}
	if gr != nil && gr.Greeting.String != "" {
		greet = gr.Greeting.String
	}

	for _, m := range *msg.NewChatMembers {
		o.log.Infof("Newbie found, send question: %s", m.FirstName)

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

		act := storage.Action{
			ChatID:    res.Chat.ID,
			Type:      "del",
			MessageID: res.MessageID,
			UserID:    m.ID,
		}
		if err := o.stor.AddToActionPool(act, time.Minute); err != nil {
			return err
		}

		err = o.stor.SetKicked(msg.Chat.ID, m.ID)
		if err != nil {
			return err
		}

		act = storage.Action{
			ChatID:    msg.Chat.ID,
			Type:      "del",
			MessageID: msg.MessageID,
			UserID:    m.ID,
		}
		if err := o.stor.AddToActionPool(act, time.Minute); err != nil {
			return err
		}

		act = storage.Action{
			ChatID: msg.Chat.ID,
			Type:   "kick",
			UserID: m.ID,
		}
		if err := o.stor.AddToActionPool(act, time.Minute); err != nil {
			return err
		}
	}

	return nil
}
