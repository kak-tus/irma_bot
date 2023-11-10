package model

import (
	"github.com/kak-tus/irma_bot/model/queries"
	"github.com/kak-tus/irma_bot/model/queries_types"
	"github.com/kak-tus/nan"
)

const defaultGreeting = `
Hello. This group has AntiSpam protection.
You must get correct answer to next question in one minute or you will be kicked.
In case of incorrect answer you can try join group after one day.
`

var defaultQuestions = queries_types.Questions{
	{
		Answers: []queries_types.Answer{
			{
				Correct: 1,
				Text:    "No",
			},
			{
				Text: "Yes",
			},
		},
		Text: "Are you a bot?",
	},
}

var defaultGroup = queries.GetGroupRow{
	BanEmojiiCount: nan.Int32(0),
	BanQuestion:    nan.Bool(true),
	BanTimeout:     nan.Int32(1),
	BanUrl:         nan.Bool(true),
	Greeting:       nan.String(defaultGreeting),
	Questions:      queries_types.QuestionsDB{Questions: defaultQuestions},
}

func (hdl *Model) GetDefaultGroup() queries.GetGroupRow {
	return defaultGroup
}
