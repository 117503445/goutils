package goutils

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// GetGitRootDir returns the root directory of the git repository
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
