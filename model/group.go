package model

import (
	"database/sql"

	"github.com/kak-tus/irma_bot/model/queries"
	"github.com/kak-tus/irma_bot/model/queries_types"
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
	BanQuestion: sql.NullBool{
		Bool:  true,
		Valid: true,
	},
	BanUrl: sql.NullBool{
		Bool:  true,
		Valid: true,
	},
	Greeting: sql.NullString{
		String: defaultGreeting,
		Valid:  true,
	},
	Questions: queries_types.QuestionsDB{Questions: defaultQuestions},
	BanTimeout: sql.NullInt32{
		Int32: 1,
		Valid: true,
	},
}

func (hdl *Model) GetDefaultGroup() queries.GetGroupRow {
	return defaultGroup
}
