package goutils

import (
	"fmt"
	"os"
	"os/exec"
)

func CMD(cwd string, command string, args ...string) error {
	var err error
	if cwd == "" {
		cwd, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
	}
	commandStr := command
	for _, arg := range args {
		commandStr += " " + arg
	}
	Logger.Debug().Str("cwd", cwd).Str("command", commandStr).Msg("Run Command")
	cmd := exec.Command(command, args...)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	err = cmd.Run()
	if err != nil {
		Logger.Error().Err(err).Str("cwd", cwd).Str("command", command).Strs("args", args).
			Msg("Failed to run command")
		return fmt.Errorf("failed to run command: %w", err)
	}
	Logger.Debug().Str("cwd", cwd).Str("command", commandStr).Msg("Run Command Done")

	return nil
}
