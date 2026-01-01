package goutils_test

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// createTestTar creates a test tar file with some content
func createTestTar(t *testing.T, tarPath string) {
	file, err := os.Create(tarPath)
	assert.NoError(t, err)
	defer file.Close()

	tarWriter := tar.NewWriter(file)
	defer tarWriter.Close()

	// Add a file to the tar
	content := "test content"
	header := &tar.Header{
		Name:    "test.txt",
		Size:    int64(len(content)),
		Mode:    0644,
		ModTime: time.Now(),
	}
	err = tarWriter.WriteHeader(header)
	assert.NoError(t, err)

	_, err = tarWriter.Write([]byte(content))
	assert.NoError(t, err)

	// Add a nested file
	nestedContent := "nested content"
	nestedHeader := &tar.Header{
		Name:    "dir/nested.txt",
		Size:    int64(len(nestedContent)),
		Mode:    0644,
		ModTime: time.Now(),
	}
	err = tarWriter.WriteHeader(nestedHeader)
	assert.NoError(t, err)

	_, err = tarWriter.Write([]byte(nestedContent))
	assert.NoError(t, err)
}

// createTestTarGz creates a test tar.gz file with some content
func createTestTarGz(t *testing.T, tarGzPath string) {
	file, err := os.Create(tarGzPath)
	assert.NoError(t, err)
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Add a file to the tar.gz
	content := "gzip test content"
	header := &tar.Header{
		Name:    "gzip_test.txt",
		Size:    int64(len(content)),
		Mode:    0644,
		ModTime: time.Now(),
	}
	err = tarWriter.WriteHeader(header)
	assert.NoError(t, err)

	_, err = tarWriter.Write([]byte(content))
	assert.NoError(t, err)
}

func TestExtractTar(t *testing.T) {
	ast := assert.New(t)

	// Create temporary directory for this test
	tempDir := t.TempDir()

	// Create test tar file
	tarPath := filepath.Join(tempDir, "test.tar")
	createTestTar(t, tarPath)

	// Extract directory
	extractDir := filepath.Join(tempDir, "extracted")
	err := goutils.Extract(context.Background(), tarPath, extractDir)
	ast.NoError(err)

	// Verify extracted files
	testFile := filepath.Join(extractDir, "test.txt")
	ast.True(goutils.FileExists(testFile))
	content, err := goutils.ReadText(testFile)
	ast.NoError(err)
	ast.Equal("test content", content)

	nestedFile := filepath.Join(extractDir, "dir", "nested.txt")
	ast.True(goutils.FileExists(nestedFile))
	nestedContent, err := goutils.ReadText(nestedFile)
	ast.NoError(err)
	ast.Equal("nested content", nestedContent)
}

func TestExtractTarGz(t *testing.T) {
	ast := assert.New(t)

	// Create temporary directory for this test
	tempDir := t.TempDir()

	// Create test tar.gz file
	tarGzPath := filepath.Join(tempDir, "test.tar.gz")
	createTestTarGz(t, tarGzPath)

	// Extract directory
	extractDir := filepath.Join(tempDir, "extracted")
	err := goutils.Extract(context.Background(), tarGzPath, extractDir)
	ast.NoError(err)

	// Verify extracted files
	testFile := filepath.Join(extractDir, "gzip_test.txt")
	ast.True(goutils.FileExists(testFile))
	content, err := goutils.ReadText(testFile)
	ast.NoError(err)
	ast.Equal("gzip test content", content)
}

func TestExtractWithConfig(t *testing.T) {
	ast := assert.New(t)

	// Create temporary directory for this test
	tempDir := t.TempDir()

	// Create test tar.gz file
	tarGzPath := filepath.Join(tempDir, "test.tar.gz")
	createTestTarGz(t, tarGzPath)

	// Extract with explicit config
	extractDir := filepath.Join(tempDir, "extracted")
	config := goutils.ExtractConfig{SrcType: "targz"}
	err := goutils.Extract(context.Background(), tarGzPath, extractDir, config)
	ast.NoError(err)

	// Verify extracted files
	testFile := filepath.Join(extractDir, "gzip_test.txt")
	ast.True(goutils.FileExists(testFile))
	content, err := goutils.ReadText(testFile)
	ast.NoError(err)
	ast.Equal("gzip test content", content)
}

func TestExtractUnsupportedType(t *testing.T) {
	ast := assert.New(t)

	// Create temporary directory for this test
	tempDir := t.TempDir()

	// Create a regular file (not tar or tar.gz)
	regularFile := filepath.Join(tempDir, "regular.txt")
	err := goutils.WriteText(regularFile, "regular content")
	ast.NoError(err)

	// Try to extract
	extractDir := filepath.Join(tempDir, "extracted")
	err = goutils.Extract(context.Background(), regularFile, extractDir)
	ast.Error(err)
	ast.Contains(err.Error(), "unsupported source type")
}

func TestExtractContextCancellation(t *testing.T) {
	ast := assert.New(t)

	// Create temporary directory for this test
	tempDir := t.TempDir()

	// Create test tar file
	tarPath := filepath.Join(tempDir, "test.tar")
	createTestTar(t, tarPath)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Try to extract
	extractDir := filepath.Join(tempDir, "extracted")
	err := goutils.Extract(ctx, tarPath, extractDir)
	ast.Error(err)
	ast.Equal(context.Canceled, err)
}
