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

type Answer struct {
	Correct int16  `json:"correct"`
	Text    string `json:"text"`
}

type Question struct {
	Answers []Answer `json:"answers"`
	Text    string   `json:"text"`
}

type Tg struct {
	BotName string
	Listen  string
	Path    string
	Proxy   string
	Token   string
	URL     string
}

type DB struct {
	DBAddr string
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
