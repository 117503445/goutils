package goutils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
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
