package telegram

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/kak-tus/irma_bot/cnf"
	"github.com/kak-tus/irma_bot/model/queries"
)

const (
	setText       = "AntiSpam protection enabled"
	greetingLimit = 1000
	questionLimit = 100
	answerLimit   = 50
)

const failTextTemplate = `
Can't parse your message.

Must be set greeting, at least one question, at least one correct answer and at least one incorrect answer.

Greeting, questions and answers has length limit.
Greeting - %d characters, question - %d, answer - %d.
`

var failText = fmt.Sprintf(failTextTemplate, greetingLimit, questionLimit, answerLimit)

func (o *InstanceObj) messageToBot(ctx context.Context, msg *tgbotapi.Message) error {
	isAdm, err := o.isAdmin(msg.Chat.ID, msg.From.ID)
	if err != nil {
		return err
	}

	if !isAdm {
		return nil
	}

	wasCommand, err := o.processCommands(ctx, msg)
	if err != nil {
		return err
	}

	if wasCommand {
		return nil
	}

	parsed, greeting, questions, err := o.parseQuestions(msg.Text)
	if err != nil {
		return err
	}

	var resp tgbotapi.MessageConfig

	if !parsed {
		resp = tgbotapi.NewMessage(msg.Chat.ID, failText)
	} else {
		resp = tgbotapi.NewMessage(msg.Chat.ID, setText)
	}

	_, err = o.bot.Send(resp)
	if err != nil {
		return err
	}

	if !parsed {
		return nil
	}

	encoded, err := jsoniter.Marshal(questions)
	if err != nil {
		return err
	}

	group := queries.CreateOrUpdateGroupQuestionsParams{
		ID:        msg.Chat.ID,
		Greeting:  sql.NullString{Valid: true, String: greeting},
		Questions: encoded,
	}

	err = o.model.Queries.CreateOrUpdateGroupQuestions(ctx, group)
	if err != nil {
		return err
	}

	return nil
}

func (o *InstanceObj) parseQuestions(txt string) (bool, string, []cnf.Question, error) {
	// @ + name + " "
	// @ + name + "\n"
	l := 1 + len(o.cnf.Telegram.BotName) + 1

	if len(txt) <= l {
		return false, "", nil, nil
	}

	// Cut bot name
	txt = txt[l:]

	var greeting string

	questions := make([]cnf.Question, 0)

	lines := strings.Split(txt, "\n")

	for _, ln := range lines {
		if strings.Contains(ln, "?") && strings.Contains(ln, ";") && strings.Contains(ln, "+") {
			pos := strings.Index(ln, "?")

			if len(ln) <= pos+1 {
				continue
			}

			question := strings.TrimSpace(ln[:pos])
			answers := ln[pos+1:]

			if len(question) > questionLimit {
				continue
			}

			answ := strings.Split(answers, ";")
			if len(answ) < 2 {
				continue
			}

			answParsed := make([]cnf.Answer, 0, len(answ))

			var hasCorrect bool

			for _, a := range answ {
				if len(a) > answerLimit {
					continue
				}

				if strings.HasPrefix(a, "+") {
					if len(a) > 1 {
						hasCorrect = true

						answParsed = append(answParsed, cnf.Answer{
							Correct: 1,
							Text:    strings.TrimSpace(a[1:]),
						})
					}
				} else {
					answParsed = append(answParsed, cnf.Answer{
						Text: strings.TrimSpace(a),
					})
				}
			}

			if !hasCorrect {
				continue
			}

			if len(answParsed) == 0 {
				continue
			}

			q := cnf.Question{
				Answers: answParsed,
				Text:    question,
			}

			questions = append(questions, q)
		} else {
			greeting += strings.TrimSpace(ln) + "\n"
		}
	}

	if greeting == "" || len(greeting) > greetingLimit || len(questions) == 0 {
		return false, "", nil, nil
	}

	return true, greeting, questions, nil
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
