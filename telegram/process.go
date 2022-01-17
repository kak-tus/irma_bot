package telegram

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/kak-tus/irma_bot/storage"
)

const usageText = `
To enable AntiSpam protection of your group:

1. Add this bot to group.
2. Grant administrator permissions to bot (this allow bot kick spammers).

By default bot uses URL protection: if newbie user send URL or forward message - bot kicks user.
You can disable or enable this protection by sending to bot:

__IRMA_BOT_NAME__ use_ban_url

or

__IRMA_BOT_NAME__ no_ban_url

Additionaly, you can add questions protection
Send message in group, format it like this:

__IRMA_BOT_NAME__
Hello. This group has AntiSpam protection.
You must get correct answer to next question in one minute or you will be kicked.
In case of incorrect answer you can try join group after one day.

Question 1?+Correct answer 1;Incorrect answer 1;Incorrect answer 2
Question 2?+Correct answer 1;+Correct answer 2;Incorrect answer 1

Disable or enable this by

__IRMA_BOT_NAME__ use_ban_question

or

__IRMA_BOT_NAME__ no_ban_question

To setup wait time before ban user send

__IRMA_BOT_NAME__ set_ban_timeout <timeout in minutes from 1 to 60>

as example

__IRMA_BOT_NAME__ set_ban_timeout 5

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
	textWithBotName := strings.ReplaceAll(usageText, botNameTemplate, o.cnf.Telegram.BotName)

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
	act := storage.Action{
		ChatID:    msg.Chat.ID,
		Type:      "del",
		MessageID: msg.MessageID,
		UserID:    int(msg.From.ID),
	}
	if err := o.stor.AddToActionPool(ctx, act, time.Second); err != nil {
		return err
	}
	act = storage.Action{
		ChatID: msg.Chat.ID,
		Type:   "kick",
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
		// if cnt > 0 && cnt <= 40 {
		return o.messageFromNewbie(ctx, msg)
	}

	if msg.NewChatMembers != nil {
		return o.newMembers(ctx, msg)
	}

	name := fmt.Sprintf("@%s", o.cnf.Telegram.BotName)

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

	userID, err := strconv.Atoi(tkns[0])
	if err != nil {
		return err
	}

	chatID, err := strconv.ParseInt(tkns[1], 10, 64)
	if err != nil {
		return err
	}

	if msg.From.ID != int64(userID) || msg.Message.Chat.ID != chatID {
		return nil
	}

	gr, err := o.model.Queries.GetGroup(ctx, chatID)
	if err != nil {
		return err
	}

	questionID, err := strconv.Atoi(tkns[2])
	if err != nil {
		return err
	}

	quest := defaultQuestions
	if len(gr.Questions) != 0 {
		err := jsoniter.Unmarshal(gr.Questions, &quest)
		if err != nil {
			return err
		}
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
