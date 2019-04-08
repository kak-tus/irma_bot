package telegram

import (
	"net/http"

	"github.com/kak-tus/irma_bot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
)

type InstanceObj struct {
	bot  *tgbotapi.BotAPI
	cnf  instanceConf
	log  *zap.SugaredLogger
	srv  *http.Server
	stor *storage.InstanceObj
}

type instanceConf struct {
	Listen    string
	NameLimit int
	Path      string
	Proxy     string
	Texts     textsConf
	Token     string
	URL       string
}

type textsConf struct {
	Usage string
}
