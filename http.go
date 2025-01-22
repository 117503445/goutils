package goutils

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func Download(url string, filePath string) error {
	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return err
	}

	client := &http.Client{}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn().Str("status", resp.Status).Msg("non-200 status code received")
	}

	return AtomicWriteFile(filePath, resp.Body)
}
