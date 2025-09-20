package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	"log/slog"

	"github.com/0x2e/fusion/api"
	"github.com/0x2e/fusion/conf"
	"github.com/0x2e/fusion/repo"
	"github.com/0x2e/fusion/service/pull"
)

func main() {
	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(l)

	config, err := conf.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return
	}

	// Reconfigure logger based on configured LOG_LEVEL.
	if conf.Debug {
		l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		slog.SetDefault(l)

		go func() {
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				slog.Error("pprof server", "error", err)
				return
			}
		}()
	} else {
		l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: config.LogLevel,
		}))
		slog.SetDefault(l)
	}
	repo.Init(config.DB)

	go pull.NewPuller(repo.NewFeed(repo.DB), repo.NewItem(repo.DB)).Run()

	api.Run(api.Params{
		Host:            config.Host,
		Port:            config.Port,
		PasswordHash:    config.PasswordHash,
		UseSecureCookie: config.SecureCookie,
		TLSCert:         config.TLSCert,
		TLSKey:          config.TLSKey,
	})
}
