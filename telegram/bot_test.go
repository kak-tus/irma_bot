package telegram

import (
	"testing"

	"github.com/kak-tus/irma_bot/cnf"
	"github.com/kak-tus/irma_bot/db"
	"github.com/stretchr/testify/assert"
)

func TestParseQuestions(t *testing.T) {
	o := InstanceObj{
		cnf: &cnf.Cnf{
			Telegram: cnf.Tg{
				BotName: "test",
				Limits: cnf.LimitsConf{
					Answer:   100,
					Greeting: 100,
					Question: 100,
				},
			},
		},
	}

	txt := `@test
	Добродошли!

	Столица Сербии?+Белград;Рашка;Сараево
	`
	parsed, questions, err := o.parseQuestions(txt)

	greet := "Добродошли!\n\n\n"

	assert.Equal(t, true, parsed, "Must be parsed")
	assert.Equal(t, &db.Group{
		Greeting: &greet,
		Questions: []cnf.Question{
			{
				Answers: []cnf.Answer{
					{Correct: 1, Text: "Белград"},
					{Correct: 0, Text: "Рашка"},
					{Correct: 0, Text: "Сараево"},
				},
				Text: "Столица Сербии",
			},
		},
	}, questions, "Must be questions")
	assert.Equal(t, nil, err, "Must not be error")
}
