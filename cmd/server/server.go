package main

import (
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

	conf.Load()
	repo.Init()

	go pull.NewPuller(repo.NewFeed(repo.DB), repo.NewItem(repo.DB)).Run()

	api.Run()
}
