package goutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// WriteJSON writes data to a file in JSON format
func WriteJSON(filename string, data interface{}) error {
	// mkdir -p
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(jsonData)
	return err
}

// ReadJSON with generic type
func ReadJSON[T any](filename string, data *T) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(data)
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
			return "", fmt.Errorf("Git repo root not found")
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
