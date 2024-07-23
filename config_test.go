package goutils

import (
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	type Config struct {
		Name string `koanf:"name"`
		Age  int    `koanf:"age"`
	}

	config := &Config{
		Name: "default-name",
		Age:  18,
	}

	ast := assert.New(t)

	loadConfig(config, []string{})

	log.Info().Interface("config", config).Msg("config loaded")

	ast.Equal("default-name", config.Name)
}
