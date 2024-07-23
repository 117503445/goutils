package goutils_test

import (
	"testing"

	"github.com/117503445/goutils"
	"github.com/stretchr/testify/assert"
)

func TestCMD(t *testing.T) {
	goutils.InitZeroLog()

	ast := assert.New(t)
	err := goutils.CMD("", "ls", "-l")
	ast.NoError(err)

	err = goutils.CMD("", "ls", "&&", "ls")
	ast.Error(err)

	err = goutils.CMD("", "ls", "&& echo 2")
	ast.Error(err)
}
