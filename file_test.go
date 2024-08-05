package goutils_test

import (
	"testing"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestGetGitRootDir(t *testing.T) {
	ast := assert.New(t)

	dir, err := goutils.GetGitRootDir()
	ast.NoError(err)
	log.Info().Str("dir", dir).Msg("git root dir")
}

func TestWriteJSON(t *testing.T) {
	ast := assert.New(t)

	data := map[string]interface{}{
		"key": "value",
	}
	filename := "test.json"
	err := goutils.WriteJSON(filename, data)
	ast.NoError(err)
}