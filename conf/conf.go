package conf

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/0x2e/fusion/auth"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const (
	Debug = false

	dotEnvFilename = ".env"
)

type Conf struct {
	Host         string
	Port         int
	PasswordHash *auth.HashedPassword
	DB           string
	SecureCookie bool
	TLSCert      string
	TLSKey       string
	LogLevel     slog.Level
}

func Load() (Conf, error) {
	if err := godotenv.Load(dotEnvFilename); err != nil {
		if !os.IsNotExist(err) {
			return Conf{}, err
		}
		slog.Warn(fmt.Sprintf("no configuration file found at %s", dotEnvFilename))
	} else {
		slog.Info(fmt.Sprintf("load configuration from %s", dotEnvFilename))
	}
	var conf struct {
		Host         string `env:"HOST" envDefault:"0.0.0.0"`
		Port         int    `env:"PORT" envDefault:"8080"`
		Password     string `env:"PASSWORD"`
		DB           string `env:"DB" envDefault:"fusion.db"`
		SecureCookie bool   `env:"SECURE_COOKIE" envDefault:"false"`
		TLSCert      string `env:"TLS_CERT"`
		TLSKey       string `env:"TLS_KEY"`
		LogLevel     string `env:"LOG_LEVEL" envDefault:"INFO"`
	}
	if err := env.Parse(&conf); err != nil {
		return Conf{}, err
	}
	slog.Debug("configuration loaded", "conf", conf)

	var pwHash *auth.HashedPassword
	if conf.Password != "" {
		hash, err := auth.HashPassword(conf.Password)
		if err != nil {
			return Conf{}, err
		}
		pwHash = &hash
	}

	if (conf.TLSCert == "") != (conf.TLSKey == "") {
		return Conf{}, errors.New("missing TLS cert or key file")
	}
	if conf.TLSCert != "" {
		conf.SecureCookie = true
	}

	// Parse log level; default to INFO on error
	level, err := ParseLogLevel(conf.LogLevel)
	if err != nil {
		slog.Warn(fmt.Sprintf("invalid LOG_LEVEL '%s', defaulting to INFO", conf.LogLevel))
		level = slog.LevelInfo
	}

	return Conf{
		Host:         conf.Host,
		Port:         conf.Port,
		PasswordHash: pwHash,
		DB:           conf.DB,
		SecureCookie: conf.SecureCookie,
		TLSCert:      conf.TLSCert,
		TLSKey:       conf.TLSKey,
		LogLevel:     level,
	}, nil
}

// ParseLogLevel validates and converts a string log level to slog.Level.
// Accepted values (case-insensitive): DEBUG, INFO, WARN, WARNING, ERROR.
// Returns an error for invalid values.
func ParseLogLevel(level string) (slog.Level, error) {
	normalized := strings.ToUpper(strings.TrimSpace(level))
	switch normalized {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "", "INFO":
		return slog.LevelInfo, nil
	case "WARN", "WARNING":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level: %s", level)
	}
}
