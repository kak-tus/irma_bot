package telegram

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/kak-tus/irma_bot/storage"
)

const usageText = `
To enable AntiSpam protection of your group:

1. Add this bot to group.
2. Grant administrator permissions (administrator, delete messages) to bot. This allow bot get
all messages, kick spammers, delete spam messages.

By default bot uses URL protection and Questions protection.

URL protection: if newbie user send URL or forward message - bot kicks user.
Questions protection: if user join to group - it asked some question from bot.

To configure bot, administrator must:
1. Start private chat with @__IRMA_BOT_NAME__.
2. Send any message to @__IRMA_BOT_NAME__ in your public chat.
3. Bot send configuration url for this public chat to private chat with administrator.

https://github.com/kak-tus/irma_bot
`

const botNameTemplate = "__IRMA_BOT_NAME__"

func (hdl *InstanceObj) process(ctx context.Context, msg tgbotapi.Update) error {
	hdl.log.Debug().Interface("msg", msg).Msg("got message")

	switch {
	case msg.Message != nil:
		return hdl.processMsg(ctx, msg.Message)
	case msg.CallbackQuery != nil:
		return hdl.processCallback(ctx, msg.CallbackQuery)
	case msg.ChatMember != nil:
		return hdl.processChatMember(ctx, msg.ChatMember)
	}

	return nil
}

func (hdl *InstanceObj) processMsg(ctx context.Context, msg *tgbotapi.Message) error {
	textWithBotName := strings.ReplaceAll(usageText, botNameTemplate, hdl.cnf.BotName)

	log := hdl.log.With().Int64("chat_id", msg.Chat.ID).
		Str("chat_name", msg.Chat.UserName).Str("chat_title", msg.Chat.Title).Logger()

	if msg.Chat.IsPrivate() {
		resp := tgbotapi.NewMessage(msg.Chat.ID, textWithBotName)

		_, err := hdl.bot.Send(resp)
		if err != nil {
			return err
		}

		return nil
	}

	// Ban users with extra long names
	// It's probably "name spammers"
	banned, err := hdl.banLongNames(log, msg.Chat.ID, msg.NewChatMembers)
	if err != nil {
		return err
	}

	if banned {
		if err := hdl.deleteMessage(msg.Chat.ID, msg.MessageID); err != nil {
			return err
		}

		return nil
	}

	banned, err = hdl.banKickPool(ctx, log, msg)
	if err != nil {
		return err
	}

	if banned {
		return nil
	}

	// Special protection from immediately added messages.
	// If user send message and newbie message is not processed yet.
	// Over some time we got this action and delete message/kick user
	// if it is in kick pool
	// Add all messages from all users
	// This is not good for huge count of messages
	// TODO
	act := storage.Action{
		ChatID:    msg.Chat.ID,
		Type:      storage.ActionTypeDelete,
		MessageID: msg.MessageID,
		UserID:    int(msg.From.ID),
	}
	if err := hdl.stor.AddToActionPool(ctx, act, time.Second); err != nil {
		return err
	}

	act = storage.Action{
		ChatID: msg.Chat.ID,
		Type:   storage.ActionTypeKick,
		UserID: int(msg.From.ID),
	}

	if err := hdl.stor.AddToActionPool(ctx, act, time.Second); err != nil {
		return err
	}

	cnt, err := hdl.stor.GetNewbieMessages(ctx, msg.Chat.ID, int(msg.From.ID))
	if err != nil {
		return err
	}

	// In case of newbie we got count >0, for ordinary user count=0
	if cnt > 0 && cnt <= 4 {
		return hdl.messageFromNewbie(ctx, log, msg)
	}

	name := fmt.Sprintf("@%s", hdl.cnf.BotName)

	if strings.HasPrefix(msg.Text, name) {
		return hdl.messageToBot(ctx, msg)
	}

	return nil
}

func (hdl *InstanceObj) processCallback(ctx context.Context, msg *tgbotapi.CallbackQuery) error {
	// UserID_ChatID_QuestionID_AnswerNum
	tkns := strings.Split(msg.Data, "_")
	if len(tkns) != 4 {
		return nil
	}

	userID, err := strconv.ParseInt(tkns[0], 10, 64)
	if err != nil {
		return err
	}

	chatID, err := strconv.ParseInt(tkns[1], 10, 64)
	if err != nil {
		return err
	}

	if msg.From.ID != userID || msg.Message.Chat.ID != chatID {
		return nil
	}

	gr, err := hdl.model.Queries.GetGroup(ctx, chatID)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	defaultGroup := hdl.model.GetDefaultGroup()

	questionID, err := strconv.Atoi(tkns[2])
	if err != nil {
		return err
	}

	quest := defaultGroup.Questions.Questions
	if len(gr.Questions.Questions) != 0 {
		quest = gr.Questions.Questions
	}

	if questionID >= len(quest) {
		return errors.New("Question id from callback greater, then questions count")
	}

	answerNum, err := strconv.Atoi(tkns[3])
	if err != nil {
		return err
	}

	if answerNum >= len(quest[questionID].Answers) {
		return errors.New("Answer num from callback greater, then answers count")
	}

	if err := hdl.deleteMessage(chatID, msg.Message.MessageID); err != nil {
		return err
	}

	if quest[questionID].Answers[answerNum].Correct == 1 {
		err := hdl.stor.DelKicked(ctx, chatID, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (hdl *InstanceObj) processChatMember(ctx context.Context, msg *tgbotapi.ChatMemberUpdated) error {
	if msg.NewChatMember.IsMember {
		return nil
	}

	if msg.NewChatMember.Status != "member" {
		return nil
	}

	if msg.NewChatMember.User == nil {
		return nil
	}

	log := hdl.log.With().Int64("chat_id", msg.Chat.ID).
		Str("chat_name", msg.Chat.UserName).Str("chat_title", msg.Chat.Title).Logger()

	// Ban users with extra long names
	// It's probably "name spammers"
	banned, err := hdl.banLongNames(log, msg.Chat.ID, []tgbotapi.User{*msg.NewChatMember.User})
	if err != nil {
		return err
	}

	if banned {
		return nil
	}

	return hdl.newMembers(ctx, log, msg.Chat.ID, []tgbotapi.User{*msg.NewChatMember.User}, 0)
}
