package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/0x2e/fusion/api"
	"github.com/0x2e/fusion/conf"
	"github.com/0x2e/fusion/pkg/logx"
	"github.com/0x2e/fusion/repo"
	"github.com/0x2e/fusion/service/pull"
)

func main() {
	defer logx.Logger.Sync()

	if conf.Debug {
		go func() {
			logx.Logger.Infoln(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	config, err := conf.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	repo.Init(config.DB)

	go pull.NewPuller(repo.NewFeed(repo.DB), repo.NewItem(repo.DB)).Run()

	api.Run(api.Params{
		Host:            config.Host,
		Port:            config.Port,
		Password:        config.Password,
		UseSecureCookie: config.SecureCookie,
		TLSCert:         config.TLSCert,
		TLSKey:          config.TLSKey,
	})
}
