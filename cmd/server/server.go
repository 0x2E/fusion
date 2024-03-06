package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/0x2e/fusion/api"
	"github.com/0x2e/fusion/conf"
	"github.com/0x2e/fusion/repo"
	"github.com/0x2e/fusion/service/pull"
)

// TODO: refactor all loggers
func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	conf.Load()
	repo.Init()

	go pull.NewPuller(repo.NewFeed(repo.DB), repo.NewItem(repo.DB)).Run()

	// TODO: pprof
	api.Run()
}
