package conf

import (
	"log/slog"
	"os"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
		wantErr  bool
	}{
		{"DEBUG", slog.LevelDebug, false},
		{"debug", slog.LevelDebug, false},
		{"INFO", slog.LevelInfo, false},
		{"", slog.LevelInfo, false},
		{"warn", slog.LevelWarn, false},
		{"WARNING", slog.LevelWarn, false},
		{"ERROR", slog.LevelError, false},
		{"TRACE", slog.LevelInfo, true},
		{"invalid", slog.LevelInfo, true},
	}

	for _, tc := range tests {
		lvl, err := ParseLogLevel(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("ParseLogLevel(%q) expected error, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Fatalf("ParseLogLevel(%q) unexpected error: %v", tc.input, err)
			}
			if lvl != tc.expected {
				t.Fatalf("ParseLogLevel(%q) expected %v, got %v", tc.input, tc.expected, lvl)
			}
		}
	}
}

func TestLoad_UsesLogLevelFromEnv(t *testing.T) {
	// Save and restore environment
	original := os.Getenv("LOG_LEVEL")
	defer func() {
		_ = os.Setenv("LOG_LEVEL", original)
	}()

	// Ensure other required envs are not interfering
	_ = os.Unsetenv("TLS_CERT")
	_ = os.Unsetenv("TLS_KEY")

	if err := os.Setenv("LOG_LEVEL", "DEBUG"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Fatalf("expected slog.LevelDebug, got %v", cfg.LogLevel)
	}

	if err := os.Setenv("LOG_LEVEL", "WARNING"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.LogLevel != slog.LevelWarn {
		t.Fatalf("expected slog.LevelWarn, got %v", cfg.LogLevel)
	}

	if err := os.Setenv("LOG_LEVEL", "invalid"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Fatalf("expected default slog.LevelInfo on invalid, got %v", cfg.LogLevel)
	}
}
