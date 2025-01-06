package conf

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const (
	Debug = false

	dotEnvFilename = ".env"
)

type Conf struct {
	Host         string `env:"HOST" envDefault:"0.0.0.0"`
	Port         int    `env:"PORT" envDefault:"8080"`
	Password     string `env:"PASSWORD"`
	DB           string `env:"DB" envDefault:"fusion.db"`
	SecureCookie bool   `env:"SECURE_COOKIE" envDefault:"false"`
	TLSCert      string `env:"TLS_CERT"`
	TLSKey       string `env:"TLS_KEY"`
}

func Load() (Conf, error) {
	if err := godotenv.Load(dotEnvFilename); err != nil {
		if !os.IsNotExist(err) {
			return Conf{}, err
		}
		log.Printf("no configuration file found at %s", dotEnvFilename)
	} else {
		log.Printf("read configuration from %s", dotEnvFilename)
	}
	var conf Conf
	if err := env.Parse(&conf); err != nil {
		panic(err)
	}
	if Debug {
		fmt.Println(conf)
	}

	if conf.Password == "" {
		return Conf{}, errors.New("password is required")
	}

	if (conf.TLSCert == "") != (conf.TLSKey == "") {
		return Conf{}, errors.New("missing TLS cert or key file")
	}
	if conf.TLSCert != "" {
		conf.SecureCookie = true
	}

	return conf, nil
}
