package telegram

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kak-tus/irma_bot/config"
	"github.com/kak-tus/irma_bot/model"
	"github.com/kak-tus/irma_bot/storage"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"
)

type InstanceObj struct {
	bot    *tgbotapi.BotAPI
	cnf    config.Tg
	lock   *sync.WaitGroup
	log    zerolog.Logger
	model  *model.Model
	oldLog *zap.SugaredLogger
	router *chi.Mux
	stop   chan bool
	stor   *storage.InstanceObj
	upd    tgbotapi.UpdatesChannel
}

type Options struct {
	Config  config.Tg
	Log     zerolog.Logger
	Model   *model.Model
	OldLog  *zap.SugaredLogger
	Router  *chi.Mux
	Storage *storage.InstanceObj
}

func NewTelegram(opts Options) (*InstanceObj, error) {
	httpClient, err := getClient(opts.Config)
	if err != nil {
		return nil, err
	}

	bot, err := tgbotapi.NewBotAPIWithClient(opts.Config.Token, tgbotapi.APIEndpoint, httpClient)
	if err != nil {
		return nil, err
	}

	inst := &InstanceObj{
		bot:    bot,
		cnf:    opts.Config,
		lock:   &sync.WaitGroup{},
		log:    opts.Log,
		model:  opts.Model,
		oldLog: opts.OldLog,
		router: opts.Router,
		stop:   make(chan bool, 1),
		stor:   opts.Storage,
	}

	return inst, nil
}

func (hdl *InstanceObj) Start() error {
	hdl.oldLog.Info("start telegram")

	webhookCnf, err := tgbotapi.NewWebhook(hdl.cnf.URL + hdl.cnf.Path)
	if err != nil {
		return err
	}

	webhookCnf.AllowedUpdates = []string{
		"message",
		"callback_query",
		"chat_member",
		"chat_join_request",
	}

	resp, err := hdl.bot.Request(webhookCnf)
	if err != nil {
		return err
	}

	hdl.oldLog.Info(resp.Description)

	hdl.upd = hdl.bot.ListenForWebhook("/" + hdl.cnf.Path)

	// HACK TODO
	// We must register our handler again in internal router
	// May be better switch telegram to other port from api?
	// o.router.Mount("/"+o.cnf.Path, http.DefaultServeMux)
	hdl.router.Route("/", func(r chi.Router) {
		r.Handle("/"+hdl.cnf.Path, http.DefaultServeMux)
	})

	hdl.processors()

	hdl.lock.Add(1)
	defer hdl.lock.Done()

	tick := time.NewTicker(time.Second * 10)

	for {
		select {
		case <-tick.C:
			err := hdl.processActions()
			if err != nil {
				hdl.oldLog.Error(err)
				continue
			}
		case <-hdl.stop:
			tick.Stop()
			return nil
		}
	}
}

func (hdl *InstanceObj) processors() {
	hdl.lock.Add(hdl.cnf.Processors)

	for i := 0; i < hdl.cnf.Processors; i++ {
		go func() {
			for {
				select {
				case <-hdl.stop:
					hdl.lock.Done()
					return
				case msg := <-hdl.upd:
					ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

					err := hdl.process(ctx, msg)
					if err != nil {
						hdl.oldLog.Error(err)
					}

					cancel()
				}
			}
		}()
	}
}

func (hdl *InstanceObj) Stop() error {
	hdl.oldLog.Info("stop telegram")

	hdl.stop <- true
	hdl.lock.Wait()

	hdl.oldLog.Info("stopped telegram")

	return nil
}

func (hdl *InstanceObj) deleteMessage(chatID int64, messageID int) error {
	del := tgbotapi.NewDeleteMessage(chatID, messageID)

	if _, err := hdl.bot.Request(del); err != nil {
		ex, ok := err.(tgbotapi.Error)
		if !(ok && ex.Message == "Bad Request: message to delete not found") {
			return err
		}

		hdl.oldLog.Warnw("Message in chat is already deleted",
			"Chat", chatID,
			"Message", messageID,
		)
	}

	return nil
}

func getClient(cnf config.Tg) (*http.Client, error) {
	httpTransport := &http.Transport{}

	if cnf.Proxy != "" {
		dialer, err := proxy.SOCKS5("tcp", cnf.Proxy, nil, proxy.Direct)
		if err != nil {
			return nil, err
		}

		httpTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			done := make(chan bool)

			var (
				con net.Conn
				err error
			)

			go func() {
				con, err = dialer.Dial(network, addr)
				done <- true
			}()

			select {
			case <-ctx.Done():
				return nil, errors.New("dial timeout")
			case <-done:
				return con, err
			}
		}
	}

	httpClient := &http.Client{Transport: httpTransport, Timeout: time.Minute}

	return httpClient, nil
}
