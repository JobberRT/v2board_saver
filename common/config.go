package common

import (
	"github.com/spf13/viper"
)

var (
	Config = &config{}
)

type (
	config struct {
		Email    string `mapstructure:"email"`
		Password string `mapstructure:"password"`
		Interval int    `mapstructure:"interval"`
		URL      string `mapstructure:"url"`
		Timeout  int    `mapstructure:"timeout"`
	}
)

func init() {
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(Config); err != nil {
		panic(err)
	}
}
