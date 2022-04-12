package configs

import (
	"log"

	"github.com/spf13/viper"
)

var config *viper.Viper

func Init() {
	var err error
	config = viper.New()
	config.SetConfigType("yaml")
	config.AddConfigPath("/etc/app/config/")
	// Uncomment below line for local development
	// config.AddConfigPath("./configs/")
	setDefaults(config)
	err = config.ReadInConfig()
	if err != nil {
		log.Fatal("Fatal error config file: ", err)
	}
}

func setDefaults(config *viper.Viper) {
	config.SetDefault("log.level", "info")
	config.SetDefault("hydra.url", "localhost:8080")
	config.SetDefault("keto.read.url", "localhost:4466")
	config.SetDefault("keto.write.url", "localhost:4467")
}

func GetConfig() *viper.Viper {
	return config
}
