package telegram

import (
	"context"
	"fmt"
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// const (
// 	greetingLimit = 1000
// 	questionLimit = 100
// 	answerLimit   = 50
// )

const (
	onlyAdminText       = "I accept messages only from admin."
	privateDisabledText = "I can't send private message with configuration url to you. Please start private chat with me."
	tokenText           = "Configuration URL %s"
)

func (o *InstanceObj) messageToBot(ctx context.Context, msg *tgbotapi.Message) error {
	isAdm, err := o.isAdmin(msg.Chat.ID, msg.From.ID)
	if err != nil {
		return err
	}

	if !isAdm {
		respMsg := tgbotapi.NewMessage(msg.Chat.ID, onlyAdminText)

		respMsg.ReplyToMessageID = msg.MessageID

		_, err = o.bot.Send(respMsg)
		if err != nil {
			return err
		}

		return nil
	}

	token, err := o.stor.NewToken(ctx, msg.Chat.ID)
	if err != nil {
		return err
	}

	u, err := url.Parse(o.cnf.URL)
	if err != nil {
		return err
	}

	qry := u.Query()
	qry.Add("token", token)

	u.RawQuery = qry.Encode()

	tokenTextWithUrl := fmt.Sprintf(tokenText, u.String())

	respMsg := tgbotapi.NewMessage(msg.From.ID, tokenTextWithUrl)

	_, err = o.bot.Request(respMsg)
	if err == nil {
		return nil
	}

	respMsg = tgbotapi.NewMessage(msg.Chat.ID, privateDisabledText)

	_, err = o.bot.Send(respMsg)
	if err != nil {
		return err
	}

	return nil
}

func (o *InstanceObj) isAdmin(chatID int64, userID int64) (bool, error) {
	req := tgbotapi.ChatAdministratorsConfig{ChatConfig: tgbotapi.ChatConfig{ChatID: chatID}}

	adms, err := o.bot.GetChatAdministrators(req)
	if err != nil {
		return false, err
	}

	for _, a := range adms {
		if a.User != nil && a.User.ID == userID {
			return true, nil
		}
	}

	return false, nil
}
