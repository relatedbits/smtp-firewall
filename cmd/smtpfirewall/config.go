package main

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

var config *viper.Viper

func init() {
	config = viper.New()
	config.AddConfigPath("./")
	config.SetConfigName("default")
	config.SetConfigType("yaml")

	if err := config.ReadInConfig(); err != nil {
		log.Fatalln("error on parsing default.yaml")
	}

	replacer := strings.NewReplacer(".", "_")
	config.SetEnvKeyReplacer(replacer)
	config.AutomaticEnv()
}
