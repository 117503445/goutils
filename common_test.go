package goutils_test

import (
	"testing"

	"github.com/117503445/goutils"
	"github.com/stretchr/testify/assert"
) 

func TestCMD(t *testing.T) {
	goutils.InitZeroLog()

	ast := assert.New(t)
	err := goutils.CMD("", "ls")
	ast.NoError(err)
}