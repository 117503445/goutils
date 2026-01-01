package goutils_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/117503445/goutils"
)

func TestGetGitRootDir(t *testing.T) {
	ast := assert.New(t)

	// This test expects to be run in a git repository
	// If not in a git repo, it should return an error
	dir, err := goutils.GetGitRootDir()
	if err != nil {
		ast.Contains(err.Error(), "not a git repository")
		log.Info().Err(err).Msg("Expected error: not in git repository")
	} else {
		ast.NotEmpty(dir)
		log.Info().Str("dir", dir).Msg("git root dir")
	}
}

func TestWriteJSON(t *testing.T) {
	ast := assert.New(t)

	// Create temporary directory for this test
	tempDir := t.TempDir()

	data := map[string]interface{}{
		"key": "value",
	}
	filename := filepath.Join(tempDir, "test.json")
	err := goutils.WriteJson(filename, data)
	ast.NoError(err)

	// Verify file was created
	ast.True(goutils.FileExists(filename))
}

func TestReadJSON(t *testing.T) {
	ast := assert.New(t)

	// Create temporary directory for this test
	tempDir := t.TempDir()

	// First write the test data
	data := map[string]interface{}{
		"key": "value",
	}
	filename := filepath.Join(tempDir, "test.json")
	err := goutils.WriteJson(filename, data)
	ast.NoError(err)

	// Now read it back
	var readData map[string]interface{}
	err = goutils.ReadJson(filename, &readData)
	ast.NoError(err)
	ast.Equal("value", readData["key"])

	type Test struct {
		Key string `json:"key"`
	}
	var test Test
	err = goutils.ReadJson(filename, &test)
	ast.NoError(err)
	ast.Equal("value", test.Key)
}

func TestReadYAML(t *testing.T) {
	ast := assert.New(t)

	// Create temporary directory for this test
	tempDir := t.TempDir()

	// First write the test data
	data := map[string]interface{}{
		"k": "v",
	}
	filename := filepath.Join(tempDir, "test.yaml")
	err := goutils.WriteYaml(filename, data)
	ast.NoError(err)

	// Now read it back
	var readData map[string]interface{}
	err = goutils.ReadYaml(filename, &readData)
	ast.NoError(err)
	ast.Equal("v", readData["k"])
}

func TestCopyFile(t *testing.T) {
	ast := assert.New(t)

	// Create temporary directory for this test
	tempDir := t.TempDir()

	// Create a test file to copy
	srcFile := filepath.Join(tempDir, "source.txt")
	err := goutils.WriteText(srcFile, "test content")
	ast.NoError(err)

	// Test copying to same directory
	destFile1 := filepath.Join(tempDir, "dest1.txt")
	err = goutils.CopyFile(srcFile, destFile1)
	ast.NoError(err)
	ast.True(goutils.FileExists(destFile1))

	// Test copying to subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	destFile2 := filepath.Join(subDir, "dest2.txt")
	err = goutils.CopyFile(srcFile, destFile2)
	ast.NoError(err)
	ast.True(goutils.FileExists(destFile2))
}

func TestCopyDir(t *testing.T) {
	ast := assert.New(t)

	// Create temporary directory for this test
	tempDir := t.TempDir()

	// Create a test directory structure to copy
	srcDir := filepath.Join(tempDir, "source")
	err := os.MkdirAll(srcDir, 0755)
	ast.NoError(err)

	// Create files in the source directory
	file1 := filepath.Join(srcDir, "file1.txt")
	err = goutils.WriteText(file1, "content1")
	ast.NoError(err)

	subDir := filepath.Join(srcDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	ast.NoError(err)

	file2 := filepath.Join(subDir, "file2.txt")
	err = goutils.WriteText(file2, "content2")
	ast.NoError(err)

	// Copy the directory
	destDir := filepath.Join(tempDir, "destination")
	err = goutils.CopyDir(srcDir, destDir)
	ast.NoError(err)

	// Verify the copy
	ast.True(goutils.DirExists(destDir))
	ast.True(goutils.FileExists(filepath.Join(destDir, "file1.txt")))
	ast.True(goutils.FileExists(filepath.Join(destDir, "subdir", "file2.txt")))
}

func TestReadText(t *testing.T) {
	ast := assert.New(t)

	// Create temporary directory for this test
	tempDir := t.TempDir()

	// Create a test file
	filename := filepath.Join(tempDir, "test.txt")
	expectedContent := "test content\nwith multiple lines"
	err := goutils.WriteText(filename, expectedContent)
	ast.NoError(err)

	// Read it back
	data, err := goutils.ReadText(filename)
	ast.NoError(err)
	ast.Equal(expectedContent, data)
}

func TestWriteText(t *testing.T) {
	ast := assert.New(t)

	// Create temporary directory for this test
	tempDir := t.TempDir()

	filename := filepath.Join(tempDir, "test.txt")
	data := "test"
	err := goutils.WriteText(filename, data)
	ast.NoError(err)

	// Verify file was created and has correct content
	ast.True(goutils.FileExists(filename))
	readData, err := goutils.ReadText(filename)
	ast.NoError(err)
	ast.Equal(data, readData)
}

func TestAtomicWriteFile(t *testing.T) {
	ast := assert.New(t)

	// Create temporary directory for this test
	tempDir := t.TempDir()

	filename := filepath.Join(tempDir, "test.txt")
	err := goutils.AtomicWriteFile(filename, strings.NewReader("test"))
	ast.NoError(err)

	// Verify file was created and has correct content
	ast.True(goutils.FileExists(filename))
	data, err := goutils.ReadText(filename)
	ast.NoError(err)
	ast.Equal("test", data)
}
