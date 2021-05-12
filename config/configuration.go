package config

import (
	"crypto/x509"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

type Config struct {
	ElasticsearchEndpoint           string         `mapstructure:"endpoint"`
	ElasticsearchAlertApiPath       string         `mapstructure:"alertAPIPath"`
	ElasticsearchRoleApiPath        string         `mapstructure:"roleAPIPath"`
	ElasticsearchUserApiPath        string         `mapstructure:"userAPIPath"`
	ElasticsearchRoleMappingApiPath string         `mapstructure:"roleMappingAPIPath"`
	ElasticsearchUsername           string         `mapstructure:"username"`
	ExtraCACertFile                 string         `mapstructure:"extraCACertFile"`
	ExtraCACert                     *x509.CertPool `mapstructure:"extraCACert"`
	ElasticsearchPassword           string         `mapstructure:"password"`
}

const (
	devConfigFile                   = "./config.yaml"
	elasticsearchEndpoint           = "ELASTICSEARCH_ENDPOINT"
	elasticsearchAlertApiPath       = "ELASTICSEARCH_ALERT_API_PATH"
	elasticsearchRoleApiPath        = "ELASTICSEARCH_ROLE_API_PATH"
	elasticsearchUserApiPath        = "ELASTICSEARCH_USER_API_PATH"
	elasticsearchRoleMappingApiPath = "ELASTICSEARCH_ROLEMAPPING_API_PATH"
	extraCACertFile                 = "EXTRA_CA_CERT_FILE"
	elasticsearchUsername           = "ELASTICSEARCH_USERNAME"
	elasticsearchPassword           = "ELASTICSEARCH_PASSWORD"
)

var (
	AppConfig    = loadConfig()
	configLogger = log.WithFields(log.Fields{
		"component": "ConfigInit",
	})
)

func loadConfig() *Config {
	viper.AutomaticEnv()
	var conf Config

	if _, err := os.Stat(devConfigFile); os.IsNotExist(err) {
		log.Println("Load configuration from environment variables")
		viper.SetDefault(elasticsearchAlertApiPath, "_opendistro/_alerting/monitors")
		viper.SetDefault(elasticsearchRoleApiPath, "_opendistro/_security/api/roles")
		viper.SetDefault(elasticsearchUserApiPath, "_opendistro/_security/api/internalusers")
		viper.SetDefault(elasticsearchRoleMappingApiPath, "_opendistro/_security/api/rolesmapping")
		viper.SetDefault(extraCACertFile, "")

		conf.ElasticsearchEndpoint = viper.GetString(elasticsearchEndpoint)
		conf.ElasticsearchAlertApiPath = viper.GetString(elasticsearchAlertApiPath)
		conf.ElasticsearchRoleApiPath = viper.GetString(elasticsearchRoleApiPath)
		conf.ElasticsearchUserApiPath = viper.GetString(elasticsearchUserApiPath)
		conf.ElasticsearchRoleMappingApiPath = viper.GetString(elasticsearchRoleMappingApiPath)
		conf.ExtraCACertFile = viper.GetString(extraCACertFile)
		conf.ElasticsearchUsername = viper.GetString(elasticsearchUsername)
		conf.ElasticsearchPassword = viper.GetString(elasticsearchPassword)

	} else {
		configLogger.Println("Load configuration from file:", devConfigFile)
		viper.SetConfigFile(devConfigFile)
		if err := viper.ReadInConfig(); err != nil {
			configLogger.Fatalf("Fatal error config file %v: %s \n", devConfigFile, err)
		}
		if err := viper.Unmarshal(&conf); err != nil {
			configLogger.Fatalf("Unable to decode into struct, %v", err)
		}
	}
	if conf.ExtraCACertFile != "" {
		conf.ExtraCACert = appendCACert(conf.ExtraCACertFile)
	}
	return &conf
}

func appendCACert(file string) *x509.CertPool {
	caCert, err := ioutil.ReadFile(file)
	if err != nil {
		configLogger.Fatalf("Unable to read file with custom CA certificates: %v", err)
	}
	// Load CA cert
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		configLogger.Fatal("Unable to add custom CA certificates to certificates pool")
	}
	return caCertPool
}
