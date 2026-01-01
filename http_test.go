package goutils_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/117503445/goutils"
)

func TestDownload(t *testing.T) {
	ast := assert.New(t)

	// Create temporary directory for this test
	tempDir := t.TempDir()

	// Test downloading a small file (we'll use a reliable test URL)
	destFile := filepath.Join(tempDir, "testfile")
	err := goutils.Download("https://httpbin.org/get", destFile)
	ast.NoError(err)

	// Verify file was created
	ast.True(goutils.FileExists(destFile))
}
