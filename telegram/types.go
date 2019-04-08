package telegram

import (
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
)

type InstanceObj struct {
	bot *tgbotapi.BotAPI
	cnf instanceConf
	log *zap.SugaredLogger
	srv *http.Server
}

type instanceConf struct {
	Listen string
	Path   string
	Proxy  string
	Texts  textsConf
	Token  string
	URL    string
}

type textsConf struct {
	Usage string
}
