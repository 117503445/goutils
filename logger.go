package goutils

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger = log.With().Str("module", "goutils").Logger()

type logOptions struct {
	NoColor bool
}

type logOption interface {
	applyTo(*logOptions) error
}

type WithNoColor struct {
}

func (w WithNoColor) applyTo(o *logOptions) error {
	o.NoColor = true
	return nil
}

func InitZeroLog(options ...logOption) {
	opt := &logOptions{
		NoColor: false,
	}

	for _, o := range options {
		err := o.applyTo(opt)
		if err != nil {
			Logger.Error().Err(err).Msg("Failed to apply log option")
		}
	}

	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05", NoColor: opt.NoColor}).Level(zerolog.DebugLevel)

	log.Logger = logger
	Logger = logger.With().Str("module", "goutils").Logger()
	CommandLogger = logger.With().Str("module", "goutils.command").Logger()
}
