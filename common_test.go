package goutils_test

import (
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/117503445/goutils"
)

func TestCommon(t *testing.T) {
	goutils.InitZeroLog(goutils.WithNoColor{})

	ast := assert.New(t)
	ast.NotEmpty(goutils.TimeStrSec())

	log.Debug().Str("TimeStrSec", goutils.TimeStrSec()).Str("TimeStrMilliSec", goutils.TimeStrMilliSec()).Msg("Time")

	log.Debug().Str("UUID4", goutils.UUID4()).Send()
	log.Debug().Str("UUID7", goutils.UUID7()).Send()

	dir, err := goutils.FindGitRepoRoot()
	ast.NoError(err)
	log.Debug().Str("GitRepoRoot", dir).Msg("GitRepoRoot")

}
