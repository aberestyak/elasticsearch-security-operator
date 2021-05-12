package logger

import (
	"os"
	"strings"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

const (
	defaultLogLevel  = "info"
	defaultLogFormat = "nested"
)

// Init initialize logger
func Init() {
	logLevelValue := os.Getenv("LOG_LEVEL")
	logLevel, err := log.ParseLevel(logLevelValue)
	if err != nil {
		log.Debugf("Wrong log level set [%v], using default [%v]", logLevelValue, defaultLogLevel)
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	logFormatValue := strings.ToLower(os.Getenv("LOG_FORMAT"))
	if logFormatValue != "text" && logFormatValue != "json" && logFormatValue != defaultLogFormat {
		log.Debugf("Wrong log format set [%v], using default [%v]", logFormatValue, defaultLogFormat)
		logFormatValue = defaultLogFormat
	}
	if logFormatValue == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else if logFormatValue == "text" {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp:          true,
			DisableLevelTruncation: true,
		})
	} else if logFormatValue == defaultLogFormat {
		log.SetFormatter(&nested.Formatter{
			HideKeys:      true,
			ShowFullLevel: true,
		})

	}
}
