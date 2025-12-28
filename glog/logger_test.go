package glog_test

import (
	"testing"

	"github.com/rs/zerolog/log"

	"github.com/117503445/goutils/glog"
)

func TestInitZeroLog(t *testing.T) {
	glog.InitZeroLog(glog.InitZeroLogConfig{})
	log.Info().Msg("InitZeroLog")
}
