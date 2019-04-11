package telegram

import (
	"fmt"
	"net/http"
	"time"

	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/applog"
	"git.aqq.me/go/app/event"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/iph0/conf"
	"github.com/kak-tus/irma_bot/settings"
	"github.com/kak-tus/irma_bot/storage"
	"golang.org/x/net/proxy"
)

var inst *InstanceObj

func init() {
	event.Init.AddHandler(
		func() error {
			cnfMap := appconf.GetConfig()["telegram"]

			var cnf instanceConf
			err := conf.Decode(cnfMap, &cnf)
			if err != nil {
				return err
			}

			httpTransport := &http.Transport{}

			if cnf.Proxy != "" {
				dialer, err := proxy.SOCKS5("tcp", cnf.Proxy, nil, proxy.Direct)
				if err != nil {
					return err
				}

				httpTransport.Dial = dialer.Dial
			}

			httpClient := &http.Client{Transport: httpTransport, Timeout: time.Minute}

			bot, err := tgbotapi.NewBotAPIWithClient(cnf.Token, httpClient)
			if err != nil {
				return err
			}

			srv := &http.Server{Addr: cnf.Listen}

			inst = &InstanceObj{
				bot:  bot,
				cnf:  cnf,
				log:  applog.GetLogger().Sugar(),
				sett: settings.Get(),
				srv:  srv,
				stor: storage.Get(),
			}

			inst.log.Info("Started telegram")

			return nil
		},
	)

	event.Stop.AddHandler(
		func() error {
			inst.log.Info("Stop telegram")

			err := inst.srv.Shutdown(nil)
			if err != nil {
				return err
			}

			inst.log.Info("Stopped telegram")
			return nil
		},
	)
}

func Get() *InstanceObj {
	return inst
}

func (o *InstanceObj) Start() error {
	res, err := o.bot.SetWebhook(tgbotapi.NewWebhook(o.cnf.URL + o.cnf.Path))
	if err != nil {
		return err
	}

	o.log.Debug(res.Description)

	upd := o.bot.ListenForWebhook("/" + o.cnf.Path)

	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	go func() {
		err := o.srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			o.log.Panic(err)
		}
	}()

	go func() {
		for {
			msg := <-upd
			err := o.process(msg)
			if err != nil {
				o.log.Error(err)
			}
		}
	}()

	return nil
}
