package goutils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// TimeStrSec returns the time format string, like 20240915.221219
func TimeStrSec() string {
	return time.Now().Format("20060102.150405")
}

// TimeStrMilliSec returns the time format string with millisecond, like 20240915.221219.123
func TimeStrMilliSec() string {
	return time.Now().Format("20060102.150405.000")
}

func UUID4() string {
	return uuid.New().String()
}

// FindGitRepoRoot
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
