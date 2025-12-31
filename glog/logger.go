package glog

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ConsoleWriterConfig struct {
	RequestId string
	DirBuild  string
	NoColor   bool
}

func NewConsoleWriter(config ...ConsoleWriterConfig) zerolog.ConsoleWriter {
	cfg := ConsoleWriterConfig{}
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.RequestId == "" {
		return zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05.000", NoColor: cfg.NoColor}
	} else {
		return zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05.000", NoColor: cfg.NoColor, FormatCaller: func(i any) string {
			var c string
			if cc, ok := i.(string); ok {
				c = cc
			}
			d := cfg.DirBuild
			if d == "unknown" {
				d = ""
			}
			if d == "" {
				if cwd, err := os.Getwd(); err == nil {
					d = cwd
				}
			}

			if d != "" {
				if rel, err := filepath.Rel(d, c); err == nil {
					c = rel
				}
			}

			if len(c) > 0 {
				c = fmt.Sprintf("[%v] %v >", cfg.RequestId, c)
			} else {
				c = fmt.Sprintf("[%v] >", cfg.RequestId)
			}
			return c
		},
		}
	}
}

func InitZeroLog(config ...InitZeroLogConfig) {
	var cfg InitZeroLogConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return file + ":" + strconv.Itoa(line)
		// if cfg.DirProject == "" {
		// 	return filepath.Base(file) + ":" + strconv.Itoa(line)
		// } else {
		// 	f := file
		// 	// 如果 file 以 cfg.DirProject 开头，则返回 file 的相对路径
		// 	if strings.HasPrefix(file, cfg.DirProject) {
		// 		f = filepath.Base(file)
		// 	}
		// 	return f + ":" + strconv.Itoa(line)
		// }
	}
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"

	logger := cfg.Logger
	if logger == nil {
		l := log.Output(NewConsoleWriter()).Level(zerolog.DebugLevel).With().Caller().Logger()
		logger = &l
	}
	log.Logger = *logger
}

type InitZeroLogConfig struct {
	Logger *zerolog.Logger
}
