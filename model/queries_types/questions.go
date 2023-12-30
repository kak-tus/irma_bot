package queries_types

import (
	"database/sql/driver"
	"fmt"

	"github.com/goccy/go-json"
)

type Answer struct {
	Correct int16  `json:"correct"`
	Text    string `json:"text"`
}

type Question struct {
	Answers []Answer `json:"answers"`
	Text    string   `json:"text"`
}

type Questions = []Question

type QuestionsDB struct {
	Questions Questions
}

func (container *QuestionsDB) Scan(value any) error {
	if value == nil {
		container.Questions = nil
		return nil
	}

	converted, ok := value.(string)
	if !ok {
		return fmt.Errorf("incorrect type '%T' for scan", value)
	}

	var vals Questions

	if err := json.Unmarshal([]byte(converted), &vals); err != nil {
		return err
	}

	container.Questions = vals

	return nil
}

func (container QuestionsDB) Value() (driver.Value, error) {
	return json.Marshal(container.Questions)
}
