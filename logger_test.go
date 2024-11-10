package goutils_test

import (
	"testing"

	"github.com/rs/zerolog/log"

	"github.com/117503445/goutils"
)

func TestInitZeroLog(t *testing.T) {
	// var err error
	// ast := assert.New(t)

	goutils.InitZeroLog()
	log.Info().Msg("InitZeroLog")

	goutils.InitZeroLog(goutils.WithNoColor{})
	log.Info().Msg("InitZeroLog WithNoColor")

	goutils.InitZeroLog(goutils.WithProduction{DirLog: "./data/logs"})
	log.Info().Msg("InitZeroLog WithProduction")
}
