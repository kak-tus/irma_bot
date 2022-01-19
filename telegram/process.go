package telegram

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func (o *InstanceObj) process(ctx context.Context, msg tgbotapi.Update) error {
	if msg.Message != nil {
		return o.processMsg(ctx, msg.Message)
	} else if msg.CallbackQuery != nil {
		return o.processCallback(ctx, msg.CallbackQuery)
	}

	return nil
}

func (o *InstanceObj) processMsg(ctx context.Context, msg *tgbotapi.Message) error {
	textWithBotName := strings.ReplaceAll(usageText, botNameTemplate, o.cnf.BotName)

	if msg.Chat.IsPrivate() {
		resp := tgbotapi.NewMessage(msg.Chat.ID, textWithBotName)

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

	banned, err = o.banKickPool(ctx, msg)
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
	if err := o.stor.AddToActionPool(ctx, act, time.Second); err != nil {
		return err
	}

	act = storage.Action{
		ChatID: msg.Chat.ID,
		Type:   storage.ActionTypeKick,
		UserID: int(msg.From.ID),
	}

	if err := o.stor.AddToActionPool(ctx, act, time.Second); err != nil {
		return err
	}

	cnt, err := o.stor.GetNewbieMessages(ctx, msg.Chat.ID, int(msg.From.ID))
	if err != nil {
		return err
	}

	// In case of newbie we got count >0, for ordinary user count=0
	if cnt > 0 && cnt <= 4 {
		return o.messageFromNewbie(ctx, msg)
	}

	if msg.NewChatMembers != nil {
		return o.newMembers(ctx, msg)
	}

	name := fmt.Sprintf("@%s", o.cnf.BotName)

	if strings.HasPrefix(msg.Text, name) {
		return o.messageToBot(ctx, msg)
	}

	return nil
}

func (o *InstanceObj) processCallback(ctx context.Context, msg *tgbotapi.CallbackQuery) error {
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

	gr, err := o.model.Queries.GetGroup(ctx, chatID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	defaultGroup := o.model.GetDefaultGroup()

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

	if err := o.deleteMessage(chatID, msg.Message.MessageID); err != nil {
		return err
	}

	if quest[questionID].Answers[answerNum].Correct == 1 {
		err := o.stor.DelKicked(ctx, chatID, userID)
		if err != nil {
			return err
		}
	}

	return nil
}
