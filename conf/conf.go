package conf

import (
	"errors"

	"github.com/spf13/viper"
)

const Debug = false

var Conf struct {
	Host     string `mapstructure:"HOST"`
	Port     int    `mapstructure:"PORT"`
	Password string `mapstructure:"PASSWORD"`
	DB       string
}

func Load() {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(&Conf); err != nil {
		panic(err)
	}
	if err := validate(); err != nil {
		panic(err)
	}
}

func validate() error {
	if Conf.Password == "" {
		return errors.New("password is required")
	}
	return nil
}
