package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-isatty"
	"github.com/patrickjmcd/reedme/internal/config"
	"github.com/patrickjmcd/reedme/internal/handler"
	"github.com/patrickjmcd/reedme/internal/pull"
	"github.com/patrickjmcd/reedme/internal/store"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	setupLogger(cfg)
	gin.SetMode(gin.ReleaseMode)

	var st store.Storer
	if cfg.DatabaseURL != "" {
		slog.Info("using PostgreSQL database", "url", maskPassword(cfg.DatabaseURL))
		st, err = store.NewPostgres(cfg.DatabaseURL)
	} else {
		slog.Info("using SQLite database", "path", cfg.DBPath)
		st, err = store.New(cfg.DBPath)
	}
	if err != nil {
		return err
	}
	defer st.Close()

	puller := pull.New(st, cfg)
	h, err := handler.New(st, cfg, puller)
	if err != nil {
		return err
	}
	r := h.SetupRouter()

	addr := ":" + strconv.Itoa(cfg.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(sigCtx)

	g.Go(func() error {
		slog.Info("starting server", "address", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		if err := puller.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		slog.Info("shutting down")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to shutdown server", "error", err)
		}

		return nil
	})

	return g.Wait()
}

func setupLogger(cfg *config.Config) {
	var level slog.Level
	switch cfg.LogLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	switch cfg.LogFormat {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	case "auto":
		if isatty.IsTerminal(os.Stdout.Fd()) {
			handler = slog.NewTextHandler(os.Stdout, opts)
		} else {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		}
	default:
		if isatty.IsTerminal(os.Stdout.Fd()) {
			handler = slog.NewTextHandler(os.Stdout, opts)
		} else {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		}
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func maskPassword(dbURL string) string {
	if dbURL == "" {
		return ""
	}
	// Simple masking for postgres:// URLs
	// Format: postgres://user:password@host:port/dbname?params
	if idx := strings.Index(dbURL, "://"); idx != -1 {
		scheme := dbURL[:idx+3]
		rest := dbURL[idx+3:]
		if atIdx := strings.Index(rest, "@"); atIdx != -1 {
			hostPart := rest[atIdx:]
			userPart := rest[:atIdx]
			if colonIdx := strings.Index(userPart, ":"); colonIdx != -1 {
				return scheme + userPart[:colonIdx] + ":***" + hostPart
			}
		}
	}
	return dbURL
}
