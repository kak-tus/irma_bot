package telegram

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/kak-tus/irma_bot/storage"
)

const defaultGreeting = `
Hello. This group has AntiSpam protection.
You must get correct answer to next question in one minute or you will be kicked.
In case of incorrect answer you can try join group after one day.
`

const defaultBanTimeout = time.Minute

var defaultQuestions = []Question{
	{
		Answers: []Answer{
			{
				Correct: 1,
				Text:    "Correct answer 1",
			},
			{
				Text: "Incorrect answer 1",
			},
			{
				Text: "Incorrect answer 2",
			},
		},
		Text: "Question 1",
	},
	{
		Answers: []Answer{
			{
				Correct: 1,
				Text:    "Correct answer 1",
			},
			{
				Correct: 1,
				Text:    "Correct answer 2",
			},
			{
				Text: "Incorrect answer 1",
			},
		},
		Text: "Question 2",
	},
}

func (o *InstanceObj) messageFromNewbie(ctx context.Context, msg *tgbotapi.Message) error {
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
		return o.stor.AddNewbieMessages(ctx, msg.Chat.ID, int(msg.From.ID))
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

	_, err := o.bot.Request(kick)
	if err != nil {
		return err
	}

	if err := o.deleteMessage(msg.Chat.ID, msg.MessageID); err != nil {
		return err
	}

	return nil
}

func (o *InstanceObj) newMembers(ctx context.Context, msg *tgbotapi.Message) error {
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

	gr, err := o.model.Queries.GetGroup(ctx, msg.Chat.ID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	for _, m := range msg.NewChatMembers {
		o.log.Infow("Newbie found, add messages",
			"User", m.FirstName,
			"Chat", msg.Chat.ID,
		)

		err := o.stor.AddNewbieMessages(ctx, msg.Chat.ID, int(m.ID))
		if err != nil {
			return err
		}
	}

	// Ban by question by default if group is not registered
	if gr.BanQuestion.Valid && !gr.BanQuestion.Bool {
		return nil
	}

	quest := defaultQuestions
	greet := defaultGreeting

	if len(gr.Questions) != 0 {
		err := jsoniter.Unmarshal(gr.Questions, &quest)
		if err != nil {
			return err
		}
	}

	if gr.Greeting.Valid {
		greet = gr.Greeting.String
	}

	for _, m := range msg.NewChatMembers {
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

		banTimeout := defaultBanTimeout
		if gr.BanTimeout.Valid {
			banTimeout = time.Duration(gr.BanTimeout.Int32) * time.Minute
		}

		act := storage.Action{
			ChatID:    res.Chat.ID,
			Type:      storage.ActionTypeDelete,
			MessageID: res.MessageID,
			UserID:    int(m.ID),
		}
		if err := o.stor.AddToActionPool(ctx, act, banTimeout); err != nil {
			return err
		}

		err = o.stor.SetKicked(ctx, msg.Chat.ID, int(m.ID), banTimeout)
		if err != nil {
			return err
		}

		act = storage.Action{
			ChatID:    msg.Chat.ID,
			Type:      storage.ActionTypeDelete,
			MessageID: msg.MessageID,
			UserID:    int(m.ID),
		}
		if err := o.stor.AddToActionPool(ctx, act, banTimeout); err != nil {
			return err
		}

		act = storage.Action{
			ChatID: msg.Chat.ID,
			Type:   storage.ActionTypeKick,
			UserID: int(m.ID),
		}
		if err := o.stor.AddToActionPool(ctx, act, banTimeout); err != nil {
			return err
		}
	}

	return nil
}
