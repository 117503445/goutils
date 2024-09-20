package goutils

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger = log.With().Str("module", "goutils").Logger()

type logOptions struct {
	NoColor bool
	Logger  *zerolog.Logger
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

type WithLogger struct {
	Logger *zerolog.Logger
}

func (w WithLogger) applyTo(o *logOptions) error {
	o.Logger = w.Logger
	return nil
}

// WithProduction is a log option, which is aimed to be used in production environment.
type WithProduction struct {
	DirLog string
}

func (w WithProduction) applyTo(o *logOptions) error {
	err := os.MkdirAll(w.DirLog, os.ModePerm)
	if err != nil {
		return err
	}

	logFilePath := fmt.Sprintf("%s/%v.jsonl", w.DirLog, TimeStrSec())

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	multiWriter := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"},
		logFile,
	)

	logger := zerolog.New(multiWriter).With().
		Timestamp().
		Caller().
		Logger()

	o.Logger = &logger
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

	var logger zerolog.Logger
	if opt.Logger == nil {
		logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05", NoColor: opt.NoColor}).Level(zerolog.DebugLevel)
	} else {
		logger = *opt.Logger
	}

	log.Logger = logger
	Logger = logger.With().Str("module", "goutils").Logger()
	CommandLogger = logger.With().Str("module", "goutils.command").Logger()
}
