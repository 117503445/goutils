package goutils_test

import (
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/117503445/goutils"
)

func TestCMD(t *testing.T) {
	goutils.InitZeroLog(goutils.WithNoColor{})

	ast := assert.New(t)
	err := goutils.CMD("", "ls", "-l")
	ast.NoError(err)

	err = goutils.CMD("", "ls", "&&", "ls")
	ast.Error(err)

	err = goutils.CMD("", "ls", "&& echo 2")
	ast.Error(err)
}

func TestExec(t *testing.T) {
	goutils.InitZeroLog(goutils.WithNoColor{})

	ast := assert.New(t)
	r, err := goutils.Exec("ls -l")
	ast.NoError(err)
	log.Debug().Str("output", r.Output).Msg("Exec")

	r, err = goutils.Exec("ls -l", goutils.WithCwd("/"))
	ast.NoError(err)
	log.Debug().Str("output", r.Output).Msg("Exec")
}
