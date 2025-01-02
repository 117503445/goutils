package goutils_test

import (
	"testing"

	"github.com/117503445/goutils"
)

func TestDownload(t *testing.T) {
	goutils.Download("https://example.com/testfile", "testfile")
}
