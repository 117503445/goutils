package goutils_test

import (
	"testing"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestCommon(t *testing.T) {
	goutils.InitZeroLog(goutils.WithNoColor{})

	ast := assert.New(t)
	ast.NotEmpty(goutils.TimeStrSec())

	log.Debug().Str("TimeStrSec", goutils.TimeStrSec()).Str("TimeStrMilliSec", goutils.TimeStrMilliSec()).Msg("Time")

	log.Debug().Str("UUID4", goutils.UUID4()).Msg("UUID4")
}
