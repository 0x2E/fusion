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

var conf struct {
	Host         string `env:"HOST" envDefault:"0.0.0.0"`
	Port         int    `env:"PORT" envDefault:"8080"`
	Password     string `env:"PASSWORD"`
	DB           string `env:"DB" envDefault:"fusion.db"`
	SecureCookie bool   `env:"SECURE_COOKIE" envDefault:"false"`
	TLSCert      string `env:"TLS_CERT"`
	TLSKey       string `env:"TLS_KEY"`
}

func Host() string       { return conf.Host }
func Port() int          { return conf.Port }
func Password() string   { return conf.Password }
func DB() string         { return conf.DB }
func SecureCookie() bool { return conf.SecureCookie }
func TLSCert() string    { return conf.TLSCert }
func TLSKey() string     { return conf.TLSKey }

func Load() {
	if err := godotenv.Load(dotEnvFilename); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		log.Printf("no configuration file found at %s", dotEnvFilename)
	} else {
		log.Printf("read configuration from %s", dotEnvFilename)
	}
	if err := env.Parse(&conf); err != nil {
		panic(err)
	}
	if err := validate(); err != nil {
		panic(err)
	}
	if Debug {
		fmt.Println(conf)
	}
}

func validate() error {
	if conf.Password == "" {
		return errors.New("password is required")
	}

	if (conf.TLSCert == "") != (conf.TLSKey == "") {
		return errors.New("missing TLS cert or key file")
	}
	if conf.TLSCert != "" {
		conf.SecureCookie = true
	}

	return nil
}
