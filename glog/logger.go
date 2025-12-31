package glog

import (
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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
		l := log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05.000", NoColor: false}).Level(zerolog.DebugLevel).With().Caller().Logger()
		logger = &l
	}
	log.Logger = *logger
}

type InitZeroLogConfig struct {
	Logger *zerolog.Logger
	// DirProject string
}
