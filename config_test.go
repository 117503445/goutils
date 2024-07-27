package goutils

import (
	"os"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	type Config struct {
		Name string `koanf:"name"`
		Age  int    `koanf:"age"`
	}
	var config *Config
	var err error
	var result *ConfigResult
	ast := assert.New(t)

	// default
	config = &Config{
		Name: "default-name",
		Age:  18,
	}
	result = loadConfig(config, []string{})
	result.Dump()
	log.Info().Interface("config", config).Msg("config loaded")
	ast.Equal("default-name", config.Name)

	// env > default
	config = &Config{
		Name: "default-name",
		Age:  18,
	}
	os.Setenv("NAME", "env-name")
	loadConfig(config, []string{})
	log.Info().Interface("config", config).Msg("config loaded")
	ast.Equal("env-name", config.Name)
	os.Unsetenv("NAME")

	// config > env > default
	config = &Config{
		Name: "default-name",
		Age:  18,
	}
	os.Setenv("NAME", "env-name")
	if err = os.WriteFile("config.toml", []byte("name = \"config-name\"\nage = 20"),
		0644); err != nil {
		panic(err)
	}
	loadConfig(config, []string{})
	log.Info().Interface("config", config).Msg("config loaded")
	ast.Equal("config-name", config.Name)
	if err = os.Remove("config.toml"); err != nil {
		panic(err)
	}
	os.Unsetenv("NAME")

	// cli > config > env > default
	config = &Config{
		Name: "default-name",
		Age:  18,
	}
	os.Setenv("NAME", "env-name")
	if err = os.WriteFile("config.toml", []byte("name = \"config-name\"\nage = 20"),
		0644); err != nil {
		panic(err)
	}
	loadConfig(config, []string{"--name", "cli-name"})
	log.Info().Interface("config", config).Msg("config loaded")
	ast.Equal("cli-name", config.Name)
	if err = os.Remove("config.toml"); err != nil {
		panic(err)
	}
	os.Unsetenv("NAME")

	// custom config file by env
	config = &Config{
		Name: "default-name",
		Age:  18,
	}
	if err = os.WriteFile("config1.toml", []byte("name = \"config-name\"\nage = 20"),
		0644); err != nil {
		panic(err)
	}
	os.Setenv("CONFIG", "config1.toml")
	loadConfig(config, []string{})
	log.Info().Interface("config", config).Msg("config loaded")
	ast.Equal("config-name", config.Name)
	if err = os.Remove("config1.toml"); err != nil {
		panic(err)
	}
	os.Unsetenv("CONFIG")

	// custom config file by cli
	config = &Config{
		Name: "default-name",
		Age:  18,
	}
	if err = os.WriteFile("config1.toml", []byte("name = \"config-name\"\nage = 20"),
		0644); err != nil {
		panic(err)
	}
	loadConfig(config, []string{"--config", "config1.toml"})
	log.Info().Interface("config", config).Msg("config loaded")
	ast.Equal("config-name", config.Name)
	if err = os.Remove("config1.toml"); err != nil {
		panic(err)
	}
}
