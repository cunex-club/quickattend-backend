package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func SetupLogger(appEnv string) zerolog.Logger {
	logger := zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()

	if appEnv == "development" {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return logger
}
