package conf

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

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

	return Conf{
		Host:         conf.Host,
		Port:         conf.Port,
		PasswordHash: pwHash,
		DB:           conf.DB,
		SecureCookie: conf.SecureCookie,
		TLSCert:      conf.TLSCert,
		TLSKey:       conf.TLSKey,
	}, nil
}
