package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func initLog() {
	lvl := os.Getenv("LOG_VERB")
	if lvl == "DEBUG" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else if lvl == "INFO" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		lvl = "ERROR"
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msgf("Log Level: %s", lvl)
}
