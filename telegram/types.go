package telegram

import (
	"net/http"
	"sync"

	"github.com/kak-tus/irma_bot/settings"
	"github.com/kak-tus/irma_bot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
)

type InstanceObj struct {
	bot  *tgbotapi.BotAPI
	cnf  instanceConf
	lock *sync.WaitGroup
	log  *zap.SugaredLogger
	sett *settings.InstanceObj
	srv  *http.Server
	stop chan bool
	stor *storage.InstanceObj
}

type instanceConf struct {
	BotName   string
	Limits    limConf
	Listen    string
	NameLimit int
	Path      string
	Proxy     string
	Texts     textsConf
	Token     string
	URL       string
}

type textsConf struct {
	Commands map[string]command
	Fail     string
	Set      string
	Usage    string
}

type command struct {
	Field string
	Text  string
	Value bool
}

type limConf struct {
	Answer   int
	Greeting int
	Question int
}
