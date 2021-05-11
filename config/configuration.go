package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	ElasticsearchEndpoint           string `mapstructure:"endpoint"`
	ElasticsearchAlertApiPath       string `mapstructure:"alertAPIPath"`
	ElasticsearchRoleApiPath        string `mapstructure:"roleAPIPath"`
	ElasticsearchUserApiPath        string `mapstructure:"userAPIPath"`
	ElasticsearchRoleMappingApiPath string `mapstructure:"roleMappingAPIPath"`
	ElasticsearchUsername           string `mapstructure:"username"`
	ElasticsearchPassword           string `mapstructure:"password"`
}

const (
	devConfigFile = "./config.yaml"
)

var (
	AppConfig = loadConfig()
)

func loadConfig() *Config {

	var conf Config

	if _, err := os.Stat(devConfigFile); os.IsNotExist(err) {
		log.Fatalf("Can't read file %v", err)
	} else {
		log.Println("Reading config file: ", devConfigFile)

		viper.SetConfigFile(devConfigFile)
		// viper.SetConfigType("json")
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Fatal error config file %v: %s \n", devConfigFile, err)
		}

		if err := viper.Unmarshal(&conf); err != nil {
			log.Fatalf("unable to decode into struct, %v", err)
		}
	}

	return &conf
}
