package telegram

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"time"
	"unicode/utf16"

	"github.com/forPelevin/gomoji"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/kak-tus/irma_bot/storage"
	"github.com/rs/zerolog"
)

const maxEmojiis = 2

func (hdl *InstanceObj) messageFromNewbie(ctx context.Context, log zerolog.Logger, msg *tgbotapi.Message) error {
	ban, err := hdl.isBanNewbie(ctx, log, msg)
	if err != nil {
		return err
	}

	if !ban {
		return hdl.stor.AddNewbieMessages(ctx, msg.Chat.ID, int(msg.From.ID))
	}

	log.Info().Str("msg", msg.Text).Msg("restricted message")

	kick := tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: msg.Chat.ID,
			UserID: msg.From.ID,
		},
		UntilDate: time.Now().In(time.UTC).AddDate(0, 0, 1).Unix(),
	}

	_, err = hdl.bot.Request(kick)
	if err != nil {
		return err
	}

	if err := hdl.deleteMessage(msg.Chat.ID, msg.MessageID); err != nil {
		return err
	}

	return nil
}

func (hdl *InstanceObj) newMembers(
	ctx context.Context,
	log zerolog.Logger,
	chatID int64,
	users []tgbotapi.User,
	initialMessageID int,
) error {
	gr, err := hdl.model.Queries.GetGroup(ctx, chatID)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	defaultGroup := hdl.model.GetDefaultGroup()

	for _, usr := range users {
		log.Info().Str("user", usr.FirstName+" "+usr.LastName).
			Str("user_name", usr.UserName).Msg("newbie found, add messages")

		err := hdl.stor.AddNewbieMessages(ctx, chatID, int(usr.ID))
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

	banTimeout := time.Duration(defaultGroup.BanTimeout.Int32) * time.Minute
	if gr.BanTimeout.Valid {
		banTimeout = time.Duration(gr.BanTimeout.Int32) * time.Minute
	}

	for _, usr := range users {
		log.Info().Str("user", usr.FirstName+" "+usr.LastName).
			Str("user_name", usr.UserName).Msg("newbie found, send question")

		qID := rand.Intn(len(quest))

		var name string
		if usr.UserName != "" {
			name = usr.UserName

			// Spammer username often has len more than 24 symbols
			// So don't display their names fully in welcome message
			if len([]rune(name)) > 24 {
				name = string([]rune(name)[0:6]) + "..."
			}
		} else {
			name = usr.FirstName

			if usr.LastName != "" {
				name += " " + usr.LastName
			}
		}

		txt := fmt.Sprintf("@%s %s\n\n%s", name, greet, quest[qID].Text)

		resp := tgbotapi.NewMessage(chatID, txt)

		btns := make([][]tgbotapi.InlineKeyboardButton, len(quest[qID].Answers))

		for i, a := range quest[qID].Answers {
			id := fmt.Sprintf("%d_%d_%d_%d", usr.ID, chatID, qID, i)
			btns[i] = []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(a.Text, id)}
		}

		resp.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(btns...)

		res, err := hdl.bot.Send(resp)
		if err != nil {
			return err
		}

		act := storage.Action{
			ChatID:    res.Chat.ID,
			Type:      storage.ActionTypeDelete,
			MessageID: res.MessageID,
			UserID:    int(usr.ID),
		}
		if err := hdl.stor.AddToActionPool(ctx, act, banTimeout); err != nil {
			return err
		}

		err = hdl.stor.SetKicked(ctx, chatID, int(usr.ID), banTimeout)
		if err != nil {
			return err
		}

		if initialMessageID != 0 {
			act = storage.Action{
				ChatID:    chatID,
				Type:      storage.ActionTypeDelete,
				MessageID: initialMessageID,
				UserID:    int(usr.ID),
			}
			if err := hdl.stor.AddToActionPool(ctx, act, banTimeout); err != nil {
				return err
			}
		}

		act = storage.Action{
			ChatID: chatID,
			Type:   storage.ActionTypeKick,
			UserID: int(usr.ID),
		}
		if err := hdl.stor.AddToActionPool(ctx, act, banTimeout); err != nil {
			return err
		}
	}

	return nil
}

func (hdl *InstanceObj) isBanNewbie(
	ctx context.Context,
	log zerolog.Logger,
	msg *tgbotapi.Message,
) (bool, error) {
	log = log.With().Str("user", msg.From.FirstName+" "+msg.From.LastName).
		Str("user_name", msg.From.UserName).Logger()

	if hdl.isBanNewbieForEntities(append(msg.Entities, msg.CaptionEntities...)...) {
		log.Info().Msg("ban for entries")
		return true, nil
	}

	if msg.ForwardFrom != nil ||
		msg.ForwardFromChat != nil ||
		msg.Sticker != nil ||
		len(msg.Photo) != 0 ||
		msg.Animation != nil ||
		msg.Audio != nil ||
		msg.Video != nil ||
		msg.VideoNote != nil ||
		msg.Voice != nil {
		log.Info().Msg("ban for media")
		return true, nil
	}

	if len(gomoji.CollectAll(msg.Text)) > maxEmojiis {
		log.Info().Msg("ban for emojii")
		return true, nil
	}

	urlsList := hdl.getURLsFromEntities(msg.Text, append(msg.Entities, msg.CaptionEntities...)...)

	if len(urlsList) != 0 {
		group, err := hdl.model.Queries.GetGroup(ctx, msg.Chat.ID)
		if err != nil && err != pgx.ErrNoRows {
			return false, err
		}

		ignore := make(map[string]struct{})

		for _, domain := range group.IgnoreDomain {
			ignore[domain] = struct{}{}
		}

		if hdl.isBanNewbieForURLs(ignore, urlsList) {
			log.Info().Msg("ban for url")
			return true, nil
		}
	}

	return false, nil
}

func (hdl *InstanceObj) isBanNewbieForEntities(
	entities ...tgbotapi.MessageEntity,
) bool {
	for _, entity := range entities {
		switch entity.Type {
		case "mention", "email":
			return true
		}
	}

	return false
}

func (hdl *InstanceObj) getURLsFromEntities(
	text string,
	entities ...tgbotapi.MessageEntity,
) []string {
	checkUrls := make([]string, 0)

	for _, entity := range entities {
		switch entity.Type {
		case "url":
			encoded16 := utf16.Encode([]rune(text))
			entityVal := encoded16[entity.Offset : entity.Offset+entity.Length]
			urlStr := string(utf16.Decode(entityVal))
			checkUrls = append(checkUrls, urlStr)
		case "text_link":
			checkUrls = append(checkUrls, entity.URL)
		}
	}

	return checkUrls
}

func (hdl *InstanceObj) isBanNewbieForURLs(
	ignore map[string]struct{},
	checkUrls []string,
) bool {
	if len(checkUrls) == 0 {
		return false
	}

	if len(ignore) == 0 {
		return true
	}

	for _, urlStr := range checkUrls {
		parsed, err := url.Parse(urlStr)
		if err != nil {
			hdl.oldLog.Errorw("can't parse url in message", "url", urlStr)

			// Ban also in case of incorrect url because it is some url
			return true
		}

		if parsed.Hostname() == "" {
			hdl.oldLog.Errorw("can't parse url in message, but no error", "url", urlStr)

			// Can't parse? Ban!
			return true
		}

		if _, ok := ignore[parsed.Hostname()]; !ok {
			return true
		}
	}

	return false
}
