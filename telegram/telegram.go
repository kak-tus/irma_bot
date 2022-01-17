package telegram

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kak-tus/irma_bot/cnf"
	"github.com/kak-tus/irma_bot/model"
	"github.com/kak-tus/irma_bot/storage"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"
)

type InstanceObj struct {
	bot   *tgbotapi.BotAPI
	cnf   *cnf.Cnf
	lock  *sync.WaitGroup
	log   *zap.SugaredLogger
	model *model.Model
	srv   *http.Server
	stop  chan bool
	stor  *storage.InstanceObj
}

func NewTelegram(log *zap.SugaredLogger) (*InstanceObj, error) {
	cnf, err := cnf.NewConf()
	if err != nil {
		return nil, err
	}

	httpClient, err := getClient(cnf)
	if err != nil {
		return nil, err
	}

	bot, err := tgbotapi.NewBotAPIWithClient(cnf.Telegram.Token, tgbotapi.APIEndpoint, httpClient)
	if err != nil {
		return nil, err
	}

	srv := &http.Server{Addr: cnf.Telegram.Listen}

	modelOpts := model.Options{
		Log: log,
		URL: cnf.DB.Addr,
	}

	model, err := model.NewModel(modelOpts)
	if err != nil {
		return nil, err
	}

	stor, err := storage.NewStorage(cnf, log)
	if err != nil {
		return nil, err
	}

	inst := &InstanceObj{
		bot:   bot,
		cnf:   cnf,
		lock:  &sync.WaitGroup{},
		log:   log,
		model: model,
		srv:   srv,
		stop:  make(chan bool, 1),
		stor:  stor,
	}

	return inst, nil
}

func (o *InstanceObj) Start() error {
	o.log.Info("start telegram")

	webhookCnf, err := tgbotapi.NewWebhook(o.cnf.Telegram.URL + o.cnf.Telegram.Path)
	if err != nil {
		return err
	}

	resp, err := o.bot.Request(webhookCnf)
	if err != nil {
		return err
	}

	o.log.Info(resp.Description)

	upd := o.bot.ListenForWebhook("/" + o.cnf.Telegram.Path)

	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	go func() {
		for {
			msg := <-upd

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

			err := o.process(ctx, msg)
			if err != nil {
				o.log.Error(err)
			}

			cancel()
		}
	}()

	tick := time.NewTicker(time.Second * 10)

	o.lock.Add(1)

	go func() {
		for {
			var stop bool

			select {
			case <-tick.C:
			case <-o.stop:
				stop = true
			}

			if stop {
				break
			}

			err := o.processActions()
			if err != nil {
				o.log.Error(err)
				continue
			}
		}

		o.lock.Done()
	}()

	err = o.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (o *InstanceObj) Stop() error {
	o.log.Info("Stop telegram")

	err := o.srv.Shutdown(context.TODO())
	if err != nil {
		return err
	}

	o.stop <- true
	o.lock.Wait()

	o.log.Info("Stopped telegram")

	return nil
}

func (o *InstanceObj) deleteMessage(chatID int64, messageID int) error {
	del := tgbotapi.NewDeleteMessage(chatID, messageID)

	if _, err := o.bot.Request(del); err != nil {
		ex, ok := err.(tgbotapi.Error)
		if ok && ex.Message == "Bad Request: message to delete not found" {
			o.log.Warnw("Message in chat is already deleted",
				"Chat", chatID,
				"Message", messageID,
			)
		} else {
			return err
		}
	}

	return nil
}

func getClient(cnf *cnf.Cnf) (*http.Client, error) {
	httpTransport := &http.Transport{}

	if cnf.Telegram.Proxy != "" {
		dialer, err := proxy.SOCKS5("tcp", cnf.Telegram.Proxy, nil, proxy.Direct)
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
