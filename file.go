package goutils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	toml "github.com/pelletier/go-toml/v2"
)

// GetGitRootDir returns the root directory of the git repository
// Deprecated: Use FindGitRepoRoot instead
func GetGitRootDir() (string, error) {
	// from execute binary file path
	f := os.Args[0]

	dir := f
	for {
		if _, err := os.Stat(dir + "/.git"); err == nil {
			return dir, nil
		}

		if dir == "/" {
			return "", errors.New("not a git repository")
		}

		dir = filepath.Dir(dir)
	}
}

// func writeFile(filename string, content []byte) error {
// 	// mkdir -p
// 	dir := filepath.Dir(filename)
// 	if err := os.MkdirAll(dir, 0755); err != nil {
// 		return err
// 	}

// 	file, err := os.Create(filename)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	_, err = file.Write(content)
// 	return err
// }

// WriteJson writes data to a file in JSON format
func WriteJson(filename string, data any) error {
	content, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	return WriteText(filename, string(content))
}

// ReadJson with generic type
func ReadJson[T any](filename string, data *T) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.NewDecoder(bytes.NewReader(content)).Decode(data)
}

func WriteYaml[T any](filename string, data T) error {
	content, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return WriteText(filename, string(content))
}

func ReadYaml[T any](filename string, data *T) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.NewDecoder(bytes.NewReader(content)).Decode(data)
}

func WriteToml[T any](filename string, data T) error {
	content, err := toml.Marshal(data)
	if err != nil {
		return err
	}

	return WriteText(filename, string(content))
}

func ReadToml[T any](filename string, data *T) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return toml.NewDecoder(file).Decode(data)
}

func ReadText(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func WriteText(filename, content string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filename, []byte(content), 0644)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	// create dst directory recursively
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// copy file mode
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// MoveFile moves a file from src to dst
func MoveFile(src, dst string) error {
	if err := CopyFile(src, dst); err != nil {
		return err
	}

	return os.Remove(src)
}

// CopyDir copies a directory from src to dst
func CopyDir(src, dst string) error {
	// create dst directory recursively
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return CopyFile(path, dstPath)
	})
}

// MoveDir moves a directory from src to dst
func MoveDir(src, dst string) error {
	if err := CopyDir(src, dst); err != nil {
		return err
	}

	return os.RemoveAll(src)
}

// FindGitRepoRoot returns the root directory of the git repository
func FindGitRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	p := wd
	for {
		if _, err := os.Stat(p + "/.git"); err == nil {
			return p, nil
		}
		if p == "/" {
			return "", fmt.Errorf("git repo root not found")
		}
		p = filepath.Dir(p)
	}
}

// PathExists returns true if the path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// AtomicWriteFile writes the content of reader to a file at path atomically.
func AtomicWriteFile(path string, reader io.Reader) error {
	// 获取目标文件所在的目录
	dir := filepath.Dir(path)

	// 定义临时文件模式，*会被替换为随机字符串
	pattern := filepath.Base(path) + ".tmp.*"

	// 创建临时文件
	tempFile, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return err
	}

	// 将reader的内容写入临时文件
	if _, err := io.Copy(tempFile, reader); err != nil {
		_ = tempFile.Close()
		return err
	}

	// 关闭临时文件
	if err := tempFile.Close(); err != nil {
		return err
	}

	// 原子性重命名临时文件为目标文件名
	if err := os.Rename(tempFile.Name(), path); err != nil {
		return err
	}

	return nil
}

type GetBuildInfoParams struct {
	Dir string
}

type BuildInfo struct {
	GitCommit  string
	GitBranch  string
	GitTag     string
	GitDirty   bool
	GitVersion string
	BuildTime  string
	BuildDir   string
}

func GetBuildInfo(ctx context.Context, params ...GetBuildInfoParams) (*BuildInfo, error) {
	param := GetBuildInfoParams{}
	if len(params) > 0 {
		param = params[0]
	}

	dir := param.Dir
	if dir == "" {
		var err error
		dir, err = FindGitRepoRoot()
		if err != nil {
			return nil, err
		}
	}

	info := &BuildInfo{}

	// Get git commit
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	if commit, err := cmd.Output(); err == nil {
		info.GitCommit = strings.TrimSpace(string(commit))
	}

	// Get git branch
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	if branch, err := cmd.Output(); err == nil {
		info.GitBranch = strings.TrimSpace(string(branch))
	}

	// Get git tag
	cmd = exec.Command("git", "tag", "--points-at", "HEAD")
	cmd.Dir = dir
	if tag, err := cmd.Output(); err == nil {
		info.GitTag = strings.TrimSpace(string(tag))
	}

	// Check if dirty
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	if status, err := cmd.Output(); err == nil {
		info.GitDirty = len(status) > 0
	}

	// Build version
	info.GitVersion = info.GitCommit
	if info.GitTag != "" {
		info.GitVersion = info.GitTag
	}
	if info.GitDirty {
		info.GitVersion = info.GitVersion + "-dirty"
	}

	// Build time
	info.BuildTime = time.Now().Format("2006-01-02 15:04:05")

	// Build dir
	info.BuildDir = dir

	return info, nil
}

type ExtractConfig struct {
	SrcType string // "tar" "targz"
}

func Extract(ctx context.Context, src string, dst string, config ...ExtractConfig) error {
	cfg := ExtractConfig{}
	if len(config) > 0 {
		cfg = config[0]
	}

	srcType := cfg.SrcType
	if srcType == "" {
		if strings.HasSuffix(src, ".tar") {
			srcType = "tar"
		} else if strings.HasSuffix(src, ".tar.gz") {
			srcType = "targz"
		} else {
			return fmt.Errorf("unsupported source type: %s", src)
		}
	}

	// Open the source file
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer file.Close()

	var tarReader *tar.Reader

	switch srcType {
	case "tar":
		tarReader = tar.NewReader(file)
	case "targz":
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		tarReader = tar.NewReader(gzipReader)
	default:
		return fmt.Errorf("unsupported source type: %s", srcType)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Extract files
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Construct the full path for the extracted file
		targetPath := filepath.Join(dst, header.Name)

		// Check for directory traversal attacks
		if !strings.HasPrefix(targetPath, filepath.Clean(dst)+string(os.PathSeparator)) &&
			targetPath != filepath.Clean(dst) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			// Create file
			dir := filepath.Dir(targetPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", targetPath, err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file %s: %w", targetPath, err)
			}
			outFile.Close()
		default:
			// Skip other types (symlinks, etc.) for now
			continue
		}
	}

	return nil
}
