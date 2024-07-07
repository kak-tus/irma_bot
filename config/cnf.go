package config

import "github.com/kelseyhightower/envconfig"

type Cnf struct {
	DB       DB
	Telegram Tg
	Storage  Stor
	Listen   string `default:":8080"`
}

type Tg struct {
	BotName    string
	Path       string
	Processors int `default:"10"`
	Proxy      string
	Token      string
	URL        string
	URLUI      string
}

type DB struct {
	Addr string
}

type Stor struct {
	RedisAddrs string
}

func NewConf() (*Cnf, error) {
	var cnf Cnf

	err := envconfig.Process("IRMA", &cnf)
	if err != nil {
		return nil, err
	}

	return &cnf, nil
}
