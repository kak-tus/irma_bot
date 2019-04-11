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

	del := tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID)

	_, err = o.bot.DeleteMessage(del)
	if err != nil {
		return err
	}

	return nil
}

func (o *InstanceObj) newMembers(msg *tgbotapi.Message) error {
	gr, err := o.sett.GetGroup(msg.Chat.ID)
	if err != nil {
		return err
	}

	if gr == nil || gr.BanURL {
		for _, m := range *msg.NewChatMembers {
			err := o.stor.AddNewbieMessages(msg.Chat.ID, m.ID)
			if err != nil {
				return err
			}
		}
	}

	if gr != nil && gr.BanQuestion && len(gr.Questions) != 0 {
		for _, m := range *msg.NewChatMembers {
			qID := rand.Intn(len(gr.Questions))

			var name string
			if m.UserName != "" {
				name = m.UserName
			} else {
				name = m.FirstName

				if m.LastName != "" {
					name += " " + m.LastName
				}
			}

			txt := fmt.Sprintf("@%s %s\n\n%s", name, gr.Greeting, gr.Questions[qID].Text)

			resp := tgbotapi.NewMessage(msg.Chat.ID, txt)

			btns := make([][]tgbotapi.InlineKeyboardButton, len(gr.Questions[qID].Answers))

			for i, a := range gr.Questions[qID].Answers {
				id := fmt.Sprintf("%d", m.ID)
				btns[i] = []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(a.Text, id)}
			}

			resp.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(btns...)

			res, err := o.bot.Send(resp)
			if err != nil {
				return err
			}

			err = o.stor.AddToActionPool(storage.Action{
				ChatID:    res.Chat.ID,
				Type:      "del",
				MessageID: res.MessageID,
			})
			if err != nil {
				return err
			}

			err = o.stor.AddToActionPool(storage.Action{
				ChatID: msg.Chat.ID,
				Type:   "kick",
				UserID: m.ID,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
