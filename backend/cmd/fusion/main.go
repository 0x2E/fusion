package main

import (
	"log"
	"strconv"

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/handler"
	"github.com/0x2E/fusion/internal/store"
)

func main() {
	cfg := config.Load()

	st, err := store.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}
	defer st.Close()

	h := handler.New(st, cfg)
	r := h.SetupRouter()

	addr := ":" + strconv.Itoa(cfg.Port)
	log.Printf("starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
