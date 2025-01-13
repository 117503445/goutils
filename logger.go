package goutils

import (
	"fmt"
	"io"
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
	Append bool // Append to existing log file, if false, it will overwrite the existing log file.
}

func (w WithProduction) applyTo(o *logOptions) error {
	if w.DirLog == "" {
		w.DirLog = "./logs"
	}

	err := os.MkdirAll(w.DirLog, os.ModePerm)
	if err != nil {
		return err
	}

	fileName := TimeStrSec()
	extList := []string{"jsonl", "log"}
	fileList := make([]io.Writer, 0)
	for _, ext := range extList {
		logFilePath := fmt.Sprintf("%s/%v.%v", w.DirLog, fileName, ext)
		// Check whether the file valid
		checkFile := func() error {
			fs, err := os.Stat(logFilePath)
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				} else {
					return nil
				}
			}
			if fs.IsDir() {
				// If the file is a directory, return an error
				return fmt.Errorf("The file path is a directory")
			}
			if !w.Append {
				// If the file exists, remove it
				if err = os.Remove(logFilePath); err != nil {
					return err
				}
			}
			return nil
		}

		if err = checkFile(); err != nil {
			return err
		}
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		fileList = append(fileList, logFile)
	}

	multiWriter := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05.000"},
		fileList[0],
		zerolog.ConsoleWriter{Out: fileList[1], TimeFormat: "2006-01-02 15:04:05.000", NoColor: true},
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

	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"

	var logger zerolog.Logger
	if opt.Logger == nil {
		logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05.000", NoColor: opt.NoColor}).Level(zerolog.DebugLevel).With().Caller().Logger()
	} else {
		logger = *opt.Logger
	}

	log.Logger = logger
	Logger = logger.With().Str("module", "goutils").Logger()
	CommandLogger = logger.With().Str("module", "goutils.command").Logger()
}
