package cnf

import (
	"github.com/iph0/conf"
	"github.com/iph0/conf/envconf"
	"github.com/iph0/conf/fileconf"
)

type Cnf struct {
	DB       DB
	Telegram Tg
	Storage  Stor
}

type Tg struct {
	BotName          string
	DefaultGreeting  string
	DefaultQuestions []Question
	Limits           limConf
	Listen           string
	NameLimit        int
	Path             string
	Proxy            string
	Texts            textsConf
	Token            string
	URL              string
}

type textsConf struct {
	Commands map[string]Command
	Fail     string
	Set      string
	Usage    string
}

type limConf struct {
	Answer   int
	Greeting int
	Question int
}

type Question struct {
	Answers []Answer `json:"answers"`
	Text    string   `json:"text"`
}

type Answer struct {
	Correct int16  `json:"correct"`
	Text    string `json:"text"`
}

type DB struct {
	DBAddr string
}

type Command struct {
	Field string
	Text  string
	Value bool
}

type Stor struct {
	RedisAddrs string
}

func NewConf() (*Cnf, error) {
	fileLdr := fileconf.NewLoader("etc", "/etc")
	envLdr := envconf.NewLoader()

	configProc := conf.NewProcessor(
		conf.ProcessorConfig{
			Loaders: map[string]conf.Loader{
				"file": fileLdr,
				"env":  envLdr,
			},
		},
	)

	configRaw, err := configProc.Load(
		"file:irma.yml",
		"env:^IRMA_",
	)

	if err != nil {
		return nil, err
	}

	var cnf Cnf
	if err := conf.Decode(configRaw["irma"], &cnf); err != nil {
		return nil, err
	}

	return &cnf, nil
}
