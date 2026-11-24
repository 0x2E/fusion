package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/handler"
	"github.com/0x2E/fusion/internal/pull"
	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-isatty"
	"golang.org/x/sync/errgroup"
)

// version is injected at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	showVersion, err := parseArgs(os.Args[1:], os.Stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		// The flag package already printed the error and usage text.
		os.Exit(2)
	}
	if showVersion {
		fmt.Println(version)
		return
	}

	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

// parseArgs reports whether the version should be printed instead of starting
// the server. It uses ContinueOnError so tests can call it and main can keep
// the flag package's default exit codes and usage output.
func parseArgs(args []string, output io.Writer) (bool, error) {
	fs := flag.NewFlagSet("fusion", flag.ContinueOnError)
	fs.SetOutput(output)

	showVersion := fs.Bool("v", false, "print version and exit")
	fs.BoolVar(showVersion, "version", false, "print version and exit")

	if err := fs.Parse(args); err != nil {
		return false, err
	}

	return *showVersion, nil
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	setupLogger(cfg)
	gin.SetMode(gin.ReleaseMode)

	st, err := store.New(cfg.DBPath)
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
		slog.Info("starting server", "address", addr, "version", version)
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
