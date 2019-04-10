package settings

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
)

type Group struct {
	BanQuestion bool
	BanURL      bool
	Greeting    string
	Questions   []Question
}

type Question struct {
	Answers []Answer
	Text    string
}

type Answer struct {
	Correct int16
	Text    string
}

func (o *InstanceObj) GetGroup(id int64) (*Group, error) {
	sql, args, err := o.sqrl.Select(
		"ban_question",
		"ban_url",
		"greeting",
		"questions",
	).
		From("public.groups").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return nil, err
	}

	tx, err := o.pool.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Commit()

	gr := &Group{}

	err = tx.QueryRow(sql, args...).Scan(
		&gr.BanQuestion,
		&gr.BanURL,
		&gr.Greeting,
		&gr.Questions,
	)

	if err != nil && err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return gr, nil
}

func (o *InstanceObj) CreateGroup(id int64, upd map[string]interface{}) error {
	cols := make([]string, 0, len(upd))

	for col := range upd {
		cols = append(cols, col)
	}
	sort.Strings(cols)

	args := make([]interface{}, 0, len(upd))
	colsConfl := make([]string, 0, len(upd))
	marks := make([]string, 0, len(upd))

	for i, col := range cols {
		args = append(args, upd[col])
		colsConfl = append(colsConfl, "EXCLUDED."+col)
		marks = append(marks, "$"+strconv.Itoa(i+2))
	}

	colsS := strings.Join(cols, ",")
	colsConflS := strings.Join(colsConfl, ",")
	marksS := strings.Join(marks, ",")

	sql := fmt.Sprintf(
		`
    INSERT INTO groups
    (id,%s) VALUES ($1,%s)
		ON CONFLICT (id) DO UPDATE SET
		(%s) = ROW(%s)
		`,
		colsS, marksS,
		colsS, colsConflS,
	)

	tx, err := o.pool.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(sql, append([]interface{}{id}, args...)...)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
