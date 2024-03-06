package conf

import (
	"errors"

	"github.com/spf13/viper"
)

var Conf struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       string
}

func Load() {
	// default conf
	Conf.Host = "0.0.0.0"
	Conf.Port = 8080
	Conf.DB = "data.db"

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
