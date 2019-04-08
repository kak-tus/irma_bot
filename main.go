package main

import (
	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/launcher"
	"github.com/iph0/conf/envconf"
	"github.com/iph0/conf/fileconf"
	"github.com/kak-tus/irma_bot/telegram"
)

func init() {
	fileLdr := fileconf.NewLoader("etc", "/etc")
	envLdr := envconf.NewLoader()

	appconf.RegisterLoader("file", fileLdr)
	appconf.RegisterLoader("env", envLdr)

	appconf.Require("file:irma.yml")
	appconf.Require("env:^IRMA_")
}

func main() {
	launcher.Run(func() error {
		err := telegram.Get().Start()
		if err != nil {
			return err
		}

		return nil
	})
}
