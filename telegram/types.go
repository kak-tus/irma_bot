package telegram

import (
	"net/http"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kak-tus/irma_bot/cnf"
	"github.com/kak-tus/irma_bot/db"
	"github.com/kak-tus/irma_bot/storage"
	"go.uber.org/zap"
)

type InstanceObj struct {
	bot  *tgbotapi.BotAPI
	cnf  *cnf.Cnf
	lock *sync.WaitGroup
	log  *zap.SugaredLogger
	db   *db.InstanceObj
	srv  *http.Server
	stop chan bool
	stor *storage.InstanceObj
}
