package telegram

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kak-tus/irma_bot/settings"
)

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

	parsed, gr, err := o.parseQuestions(msg.Text)
	if err != nil {
		return err
	}

	var resp tgbotapi.MessageConfig

	if !parsed {
		resp = tgbotapi.NewMessage(msg.Chat.ID, o.cnf.Texts.Fail)
	} else {
		resp = tgbotapi.NewMessage(msg.Chat.ID, o.cnf.Texts.Set)
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

	err = o.sett.CreateGroup(msg.Chat.ID, toCr)
	if err != nil {
		return err
	}

	return nil
}

func (o *InstanceObj) parseQuestions(txt string) (bool, *settings.Group, error) {
	// @ + name + " "
	l := 1 + len(o.cnf.BotName) + 1

	if len(txt) <= l {
		return false, nil, nil
	}

	txt = txt[l:]

	var greeting string
	qst := make([]settings.Question, 0)

	lines := strings.Split(txt, "\n")
	for _, ln := range lines {
		if strings.Contains(ln, "?") && strings.Contains(ln, ";") && strings.Contains(ln, "+") {
			pos := strings.Index(ln, "?")

			if len(ln) <= pos+1 {
				continue
			}

			question := strings.TrimSpace(ln[:pos])
			answers := ln[pos+1:]

			if len(question) > o.cnf.Limits.Question {
				continue
			}

			answ := strings.Split(answers, ";")
			if len(answ) < 2 {
				continue
			}

			answParsed := make([]settings.Answer, 0, len(answ))
			var hasCorrect bool

			for _, a := range answ {
				if len(a) > o.cnf.Limits.Answer {
					continue
				}

				if strings.HasPrefix(a, "+") {
					if len(a) > 1 {
						hasCorrect = true
						answParsed = append(answParsed, settings.Answer{
							Correct: 1,
							Text:    strings.TrimSpace(a[1:]),
						})
					}
				} else {
					answParsed = append(answParsed, settings.Answer{
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

			q := settings.Question{
				Answers: answParsed,
				Text:    question,
			}

			qst = append(qst, q)
		} else {
			greeting += strings.TrimSpace(ln) + "\n"
		}
	}

	if greeting == "" || len(greeting) > o.cnf.Limits.Greeting || len(qst) == 0 {
		return false, nil, nil
	}

	gr := &settings.Group{
		Greeting:  greeting,
		Questions: qst,
	}

	return true, gr, nil
}