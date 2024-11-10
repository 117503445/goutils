package goutils_test

import (
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/117503445/goutils"
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

func TestReadJSON(t *testing.T) {
	ast := assert.New(t)

	filename := "test.json"
	var data map[string]interface{}
	err := goutils.ReadJSON(filename, &data)
	ast.NoError(err)
	ast.Equal("value", data["key"])

	type Test struct {
		Key string `json:"key"`
	}
	var test Test
	err = goutils.ReadJSON(filename, &test)
	ast.NoError(err)
	ast.Equal("value", test.Key)
}

func TestCopyFile(t *testing.T) {
	goutils.CopyFile("go.mod", "go.mod.bak")
	goutils.CopyFile("go.mod", "1/go.mod.bak")
}

func TestCopyDir(t *testing.T) {
	goutils.CopyDir("data", "data1")
}

func TestReadText(t *testing.T) {
	ast := assert.New(t)

	filename := "go.mod"
	data, err := goutils.ReadText(filename)
	ast.NoError(err)
	log.Info().Str("data", data).Msg("ReadText")
}

func TestWriteText(t *testing.T) {
	ast := assert.New(t)

	filename := "test.txt"
	data := "test"
	err := goutils.WriteText(filename, data)
	ast.NoError(err)
}
