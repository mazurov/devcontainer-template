package logger

import (
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	// Default configuration
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

// GetLogger returns the configured logger instance
func GetLogger() *logrus.Logger {
	return log
}

// SetLevel sets the logging level
func SetLevel(level string) error {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	return nil
}
