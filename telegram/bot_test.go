package telegram

import (
	"testing"

	"github.com/kak-tus/irma_bot/cnf"
	"github.com/kak-tus/irma_bot/model/queries_types"
	"github.com/stretchr/testify/require"
)

func TestParseQuestions(t *testing.T) {
	o := InstanceObj{
		cnf: cnf.Tg{
			BotName: "test",
		},
	}

	txt := `@test
	Добродошли!

	Столица Сербии?+Белград;Рашка;Сараево
	`
	parsed, greeting, questions, err := o.parseQuestions(txt)

	require.NoError(t, err, "must not be error")

	greet := "Добродошли!\n\n\n"

	require.True(t, parsed, "must be parsed")
	require.Equal(t, greet, greeting)

	require.Equal(t, []queries_types.Question{
		{
			Answers: []queries_types.Answer{
				{Correct: 1, Text: "Белград"},
				{Correct: 0, Text: "Рашка"},
				{Correct: 0, Text: "Сараево"},
			},
			Text: "Столица Сербии",
		},
	}, questions, "Must be questions")
}
