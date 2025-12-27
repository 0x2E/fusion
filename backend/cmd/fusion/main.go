package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/handler"
	"github.com/0x2E/fusion/internal/pull"
	"github.com/0x2E/fusion/internal/store"
)

func main() {
	cfg := config.Load()

	st, err := store.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}
	defer st.Close()

	// Start pull service in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	puller := pull.New(st, cfg)
	go func() {
		if err := puller.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("pull service error: %v", err)
		}
	}()

	h := handler.New(st, cfg, puller)
	r := h.SetupRouter()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := ":" + strconv.Itoa(cfg.Port)
		log.Printf("starting server on %s", addr)
		if err := r.Run(addr); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down gracefully...")
	cancel()
}
