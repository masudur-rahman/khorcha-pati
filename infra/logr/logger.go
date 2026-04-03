package logr

import (
	"log"
	"os"

	"go.uber.org/zap"
)

// Logger is an interface to print logs
type Logger interface {
	Infof(format string, v ...interface{})
	Infow(msg string, keysAndValues ...interface{})

	Warnf(format string, v ...interface{})
	Warnw(msg string, keysAndValues ...interface{})

	Errorf(format string, args ...interface{})
	Errorw(msg string, keysAndValues ...interface{})

	Fatalf(format string, v ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
}

var DefaultLogger Logger

func init() {
	var (
		logger *zap.Logger
		err    error
	)
	if os.Getenv("ENV") == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatal(err)
	}

	DefaultLogger = logger.Sugar()
}
