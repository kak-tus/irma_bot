package telegram

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kak-tus/irma_bot/storage"
)

func (o *InstanceObj) process(msg tgbotapi.Update) error {
	if msg.Message != nil {
		return o.processMsg(msg.Message)
	} else if msg.CallbackQuery != nil {
		return o.processCallback(msg.CallbackQuery)
	}

	return nil
}

func (o *InstanceObj) processMsg(msg *tgbotapi.Message) error {
	if msg.Chat.IsPrivate() {
		resp := tgbotapi.NewMessage(msg.Chat.ID, o.cnf.Texts.Usage)
		_, err := o.bot.Send(resp)
		if err != nil {
			return err
		}

		return nil
	}

	// Ban users with extra long names
	// It's probably "name spammers"
	banned, err := o.banLongNames(msg)
	if err != nil {
		return err
	}
	if banned {
		return nil
	}

	banned, err = o.banKickPool(msg)
	if err != nil {
		return err
	}
	if banned {
		return nil
	}

	cnt, err := o.stor.GetNewbieMessages(msg.Chat.ID, msg.From.ID)
	if err != nil {
		return err
	}

	// In case of newbie we got count >0, for ordinary user count=0
	if cnt > 0 && cnt <= 4 {
		// if cnt > 0 && cnt <= 40 {
		return o.messageFromNewbie(msg)
	}

	if msg.NewChatMembers != nil {
		return o.newMembers(msg)
	}

	name := fmt.Sprintf("@%s", o.cnf.BotName)

	if strings.HasPrefix(msg.Text, name) {
		return o.messageToBot(msg)
	}

	return nil
}

func (o *InstanceObj) processCallback(msg *tgbotapi.CallbackQuery) error {
	return nil
}

func (o *InstanceObj) banLongNames(msg *tgbotapi.Message) (bool, error) {
	if msg.NewChatMembers == nil {
		return false, nil
	}

	var toDel bool

	for _, u := range *msg.NewChatMembers {
		if len(u.FirstName) >= o.cnf.NameLimit || len(u.LastName) >= o.cnf.NameLimit {
			o.log.Infof("Ban long name: %s", u.FirstName)
			toDel = true
			break
		}
	}

	if !toDel {
		return false, nil
	}

	for _, u := range *msg.NewChatMembers {
		kick := tgbotapi.KickChatMemberConfig{
			ChatMemberConfig: tgbotapi.ChatMemberConfig{
				ChatID: msg.Chat.ID,
				UserID: u.ID,
			},
			UntilDate: time.Now().In(time.UTC).AddDate(0, 0, 1).Unix(),
		}

		_, err := o.bot.KickChatMember(kick)
		if err != nil {
			return true, err
		}
	}

	del := tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID)

	_, err := o.bot.DeleteMessage(del)
	if err != nil {
		return true, err
	}

	return true, nil
}

func (o *InstanceObj) banKickPool(msg *tgbotapi.Message) (bool, error) {
	kicked, err := o.stor.IsKicked(msg.Chat.ID, msg.From.ID)
	if err != nil {
		return false, err
	}

	if !kicked {
		return false, err
	}

	o.log.Infof("User found in kick pool: %s", msg.From.FirstName)

	kick := tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: msg.Chat.ID,
			UserID: msg.From.ID,
		},
		UntilDate: time.Now().In(time.UTC).AddDate(0, 0, 1).Unix(),
	}

	_, err = o.bot.KickChatMember(kick)
	if err != nil {
		return true, err
	}

	return true, nil
}

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

func (o *InstanceObj) messageToBot(msg *tgbotapi.Message) error {
	adms, err := o.bot.GetChatAdministrators(tgbotapi.ChatConfig{ChatID: msg.Chat.ID})
	if err != nil {
		return err
	}

	var foundAdm bool

	for _, a := range adms {
		if a.User != nil && a.User.ID == msg.From.ID {
			foundAdm = true
			break
		}
	}

	if !foundAdm {
		return nil
	}

	for k, v := range o.cnf.Texts.Commands {
		if !strings.Contains(msg.Text, k) {
			continue
		}

		o.log.Debugf("Command %s", k)

		err := o.sett.CreateGroup(msg.Chat.ID, map[string]interface{}{v.Field: v.Value})
		if err != nil {
			return err
		}

		resp := tgbotapi.NewMessage(msg.Chat.ID, v.Text)

		_, err = o.bot.Send(resp)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}
