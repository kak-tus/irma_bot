// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0

package queries

import (
	"database/sql"

	"github.com/kak-tus/irma_bot/model/queries_types"
)

// Group data.
type Group struct {
	// id.
	ID int64
	// Greeting.
	Greeting sql.NullString
	// Questions and answers.
	Questions queries_types.QuestionsDB
	// Ban by postings urls and forwards.
	BanUrl sql.NullBool
	// Ban by question.
	BanQuestion sql.NullBool
	// Ban delay after question
	BanTimeout sql.NullInt32
}
