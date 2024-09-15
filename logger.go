package goutils

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger = log.With().Str("module", "goutils").Logger()

func InitZeroLog() {
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}).Level(zerolog.DebugLevel)

	log.Logger = logger
}
