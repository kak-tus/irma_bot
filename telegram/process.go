package telegram

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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

	if msg.From.ID != userID || msg.Message.Chat.ID != chatID {
		return nil
	}

	gr, err := o.sett.GetGroup(chatID)
	if err != nil {
		return err
	}

	if gr == nil {
		return errors.New("No group found for callback")
	}

	questionID, err := strconv.Atoi(tkns[2])
	if err != nil {
		return err
	}

	if questionID >= len(gr.Questions) {
		return errors.New("Question id from callback greater, then questions count")
	}

	answerNum, err := strconv.Atoi(tkns[3])
	if err != nil {
		return err
	}

	if answerNum >= len(gr.Questions[questionID].Answers) {
		return errors.New("Answer num from callback greater, then answers count")
	}

	del := tgbotapi.NewDeleteMessage(chatID, msg.Message.MessageID)
	_, err = o.bot.Send(del)
	if err != nil {
		return err
	}

	if gr.Questions[questionID].Answers[answerNum].Correct == 1 {
		err := o.stor.DelKicked(chatID, userID)
		if err != nil {
			return err
		}
	}

	return nil
}
