package api

import (
	"database/sql"
	"net/http"

	"github.com/goccy/go-json"
)

func (hdl *API) GetGroup(w http.ResponseWriter, r *http.Request, id GroupID) {
	group, err := hdl.model.Queries.GetGroup(r.Context(), int64(id))
	if err != nil && err != sql.ErrNoRows {
		hdl.errorInternal(w, err, "not found")
		return
	}

	defaultGroup := hdl.model.GetDefaultGroup()

	respGroup := GetGroupResponse{
		BanQuestion: defaultGroup.BanQuestion.Bool,
		BanTimeout:  defaultGroup.BanTimeout.Int32,
		BanUrl:      defaultGroup.BanUrl.Bool,
		Greeting:    defaultGroup.Greeting.String,
		Id:          int64(id),
	}

	if group.BanQuestion.Valid {
		respGroup.BanQuestion = group.BanQuestion.Bool
	}

	if group.BanTimeout.Valid {
		respGroup.BanTimeout = group.BanTimeout.Int32
	}

	if group.BanUrl.Valid {
		respGroup.BanUrl = group.BanUrl.Bool
	}

	if group.Greeting.Valid {
		respGroup.Greeting = group.Greeting.String
	}

	questions := defaultGroup.Questions.Questions
	if len(group.Questions.Questions) != 0 {
		questions = group.Questions.Questions
	}

	for _, question := range questions {
		respAnswers := make([]Answer, len(question.Answers))

		for i, answer := range question.Answers {
			respAnswer := Answer{
				Text: answer.Text,
			}

			if answer.Correct == 1 {
				val := true
				respAnswer.Correct = &val
			}

			respAnswers[i] = respAnswer
		}

		respQuestion := Question{
			Answers: respAnswers,
			Text:    question.Text,
		}

		respGroup.Questions = append(respGroup.Questions, respQuestion)
	}

	_ = json.NewEncoder(w).Encode(respGroup)
}
