package telegram

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kak-tus/irma_bot/cnf"
	"github.com/kak-tus/irma_bot/db"
)

func (o *InstanceObj) messageToBot(msg *tgbotapi.Message) error {
	isAdm, err := o.isAdmin(msg.Chat.ID, msg.From.ID)
	if err != nil {
		return err
	}

	if !isAdm {
		return nil
	}

	for k, v := range o.cnf.Telegram.Texts.Commands {
		if !strings.Contains(msg.Text, k) {
			continue
		}

		o.log.Infof("Command %s", k)

		var val interface{}

		if v.ValueFromText {
			idx := strings.Index(msg.Text, k)

			// already checked on contains
			toParse := strings.TrimSpace(msg.Text[idx+len(k):])

			parsed, err := strconv.Atoi(toParse)
			if err != nil {
				o.log.Debugf("parse of %s failed with %v", toParse, err)
				continue
			}

			if parsed < v.Minimum || parsed > v.Maximum {
				continue
			}

			val = parsed
		} else {
			val = v.Value
		}

		err := o.db.CreateGroup(msg.Chat.ID, map[string]interface{}{v.Field: val})
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

	parsed, gr, err := o.parseQuestions(msg.Text)
	if err != nil {
		return err
	}

	var resp tgbotapi.MessageConfig

	if !parsed {
		resp = tgbotapi.NewMessage(msg.Chat.ID, o.cnf.Telegram.Texts.Fail)
	} else {
		resp = tgbotapi.NewMessage(msg.Chat.ID, o.cnf.Telegram.Texts.Set)
	}

	_, err = o.bot.Send(resp)
	if err != nil {
		return err
	}

	if !parsed {
		return nil
	}

	toCr := map[string]interface{}{
		"greeting":  gr.Greeting,
		"questions": gr.Questions,
	}

	err = o.db.CreateGroup(msg.Chat.ID, toCr)
	if err != nil {
		return err
	}

	return nil
}

func (o *InstanceObj) parseQuestions(txt string) (bool, *db.Group, error) {
	// @ + name + " "
	// @ + name + "\n"
	l := 1 + len(o.cnf.Telegram.BotName) + 1

	if len(txt) <= l {
		return false, nil, nil
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

			if len(question) > o.cnf.Telegram.Limits.Question {
				continue
			}

			answ := strings.Split(answers, ";")
			if len(answ) < 2 {
				continue
			}

			answParsed := make([]cnf.Answer, 0, len(answ))
			var hasCorrect bool

			for _, a := range answ {
				if len(a) > o.cnf.Telegram.Limits.Answer {
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

	if greeting == "" || len(greeting) > o.cnf.Telegram.Limits.Greeting || len(questions) == 0 {
		return false, nil, nil
	}

	gr := &db.Group{
		Greeting:  &greeting,
		Questions: questions,
	}

	return true, gr, nil
}

func (o *InstanceObj) isAdmin(chatID int64, userID int) (bool, error) {
	adms, err := o.bot.GetChatAdministrators(tgbotapi.ChatConfig{ChatID: chatID})
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
