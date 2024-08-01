package conf

import (
	"errors"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const Debug = false

var Conf struct {
	Host     string `env:"HOST" envDefault:"0.0.0.0"`
	Port     int    `env:"PORT" envDefault:"8080"`
	Password string `env:"PASSWORD"`
	DB       string `env:"DB" envDefault:"fusion.db"`

	SecureCookie bool   `env:"SECURE_COOKIE" envDefault:"false"`
	TLSCert      string `env:"TLS_CERT"`
	TLSKey       string `env:"TLS_KEY"`
}

func Load() {
	if err := godotenv.Load(".env"); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		fmt.Println("cannot find .env, skip")
	}
	if err := env.Parse(&Conf); err != nil {
		panic(err)
	}
	if err := validate(); err != nil {
		panic(err)
	}
	if Debug {
		fmt.Println(Conf)
	}
}

func validate() error {
	if Conf.Password == "" {
		return errors.New("password is required")
	}

	if (Conf.TLSCert == "") != (Conf.TLSKey == "") {
		return errors.New("missing TLS cert or key file")
	}
	if Conf.TLSCert != "" {
		Conf.SecureCookie = true
	}

	return nil
}
