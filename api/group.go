package api

import (
	"database/sql"
	"io"
	"net/http"
	"net/url"

	"github.com/goccy/go-json"
	"github.com/kak-tus/irma_bot/model/queries"
	"github.com/kak-tus/irma_bot/model/queries_types"
	"github.com/kak-tus/nan"
)

func (hdl *API) GetGroup(w http.ResponseWriter, r *http.Request, params GetGroupParams) {
	data, err := hdl.storage.GetTokenData(r.Context(), string(params.Token))
	if err != nil {
		hdl.errorInternal(w, err, "get group failed")
		return
	}

	if data.ChatID == 0 {
		hdl.errorNotFound(w, err, "zero chat id")
		return
	}

	group, err := hdl.model.Queries.GetGroup(r.Context(), data.ChatID)
	if err != nil && err != sql.ErrNoRows {
		hdl.errorNotFound(w, err, "not found")
		return
	}

	// TODO move default values to DB
	defaultGroup := hdl.model.GetDefaultGroup()

	respGroup := GetGroupResponse{
		BanQuestion:    defaultGroup.BanQuestion.Bool,
		BanTimeout:     defaultGroup.BanTimeout.Int32,
		BanUrl:         defaultGroup.BanUrl.Bool,
		Greeting:       defaultGroup.Greeting.String,
		Id:             data.ChatID,
		BanEmojiiCount: defaultGroup.BanEmojiiCount.Addr(),
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

	if len(group.IgnoreDomain) != 0 {
		respGroup.IgnoreDomains = &group.IgnoreDomain
	}

	if group.BanEmojiiCount.Valid {
		respGroup.BanEmojiiCount = group.BanEmojiiCount.Addr()
	}

	_ = json.NewEncoder(w).Encode(respGroup)
}

func (hdl *API) SaveGroup(w http.ResponseWriter, r *http.Request, params SaveGroupParams) {
	data, err := hdl.storage.GetTokenData(r.Context(), string(params.Token))
	if err != nil {
		hdl.errorInternal(w, err, "internal error")
		return
	}

	if data.ChatID == 0 {
		hdl.errorNotFound(w, err, "not found")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		hdl.errorInternal(w, err, "internal error")
		return
	}

	var group Group

	err = json.Unmarshal(body, &group)
	if err != nil {
		hdl.errorInternal(w, err, "internal error")
		return
	}

	questions := make([]queries_types.Question, len(group.Questions))

	for questionIdx, question := range group.Questions {
		answers := make([]queries_types.Answer, len(question.Answers))

		for answerIdx, answer := range question.Answers {
			answers[answerIdx] = queries_types.Answer{
				Text: answer.Text,
			}

			if answer.Correct != nil && *answer.Correct {
				answers[answerIdx].Correct = 1
			}
		}

		questions[questionIdx] = queries_types.Question{
			Answers: answers,
			Text:    question.Text,
		}
	}

	updateParams := queries.CreateOrUpdateGroupParams{
		ID:       data.ChatID,
		Greeting: nan.String(group.Greeting),
		Questions: queries_types.QuestionsDB{
			Questions: questions,
		},
		BanUrl:         nan.Bool(group.BanUrl),
		BanQuestion:    nan.Bool(group.BanQuestion),
		BanTimeout:     nan.Int32(group.BanTimeout),
		BanEmojiiCount: nan.Int32Addr(group.BanEmojiiCount),
	}

	if group.IgnoreDomains != nil {
		ignore := make([]string, 0, len(*group.IgnoreDomains))

		for _, domain := range *group.IgnoreDomains {
			parsed, err := url.Parse("http://" + domain)
			if err != nil {
				hdl.errorInternal(w, err, "invalid domain")
				return
			}

			ignore = append(ignore, parsed.Hostname())
		}

		updateParams.IgnoreDomain = ignore
	}

	err = hdl.model.Queries.CreateOrUpdateGroup(r.Context(), updateParams)
	if err != nil {
		hdl.errorInternal(w, err, "internal error")
		return
	}

	_ = json.NewEncoder(w).Encode(SaveGroupResponse{})
}
