package gexec

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Command creates a new exec.Cmd instance by splitting the command string using cmd.split(" ").
// Important Note: This method poses significant security risks, particularly related to parameter handling errors.
// If the input string contains spaces but should be treated as a single argument (e.g., a path like "C:\Program Files\MyApp"),
// directly using strings.Split(cmd, " ") may result in incorrect parsing of arguments, leading to command execution failures or unexpected behavior.
// More critically, if unverified user inputs are used to construct and execute commands, it can lead to command injection attacks,
// where attackers can execute arbitrary system commands through specially crafted inputs.
// To enhance security, ensure all inputs are rigorously validated and consider using safer methods to pass arguments to the exec.Command function.
func Command(cmd string) *exec.Cmd {
	return Commands(strings.Split(cmd, " "))
}

// Commands creates a new exec.Cmd instance using the provided command and its arguments list.
// Important: Ensure all inputs have been properly validated and sanitized to avoid any security issues arising from improper parameter handling.
func Commands(cmds []string) *exec.Cmd {
	cmd := exec.Command(cmds[0], cmds[1:]...)
	return cmd
}

func SetEnvs( envs map[string]string,cmd *exec.Cmd) *exec.Cmd {
	for k, v := range envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	return cmd
}

func SetPwd(pwd string,cmd *exec.Cmd ) *exec.Cmd {
	cmd.Dir = pwd
	return cmd
}

type RunCfg struct {
	DisableLog bool
	Writers    []io.Writer
}

func Run(cmd *exec.Cmd, cfg ...*RunCfg) (string, error) {
	config := &RunCfg{}
	if len(cfg) == 1 {
		if cfg[0] == nil {
			log.Fatal().Msg("Run config is nil")
		}
		config = cfg[0]
	} else if len(cfg) > 1 {
		log.Fatal().Msg("Run only support one config")
	}

	formatDuration := func(d time.Duration) string {
		// 将 duration 转换为秒
		sec := d.Seconds()

		// 确定合适的单位和数值范围
		if sec < 1 {
			ms := d.Milliseconds() // 毫秒
			return fmt.Sprintf("%dms", ms)
		} else if sec >= 1 && sec < 60 {
			return fmt.Sprintf("%.3fs", sec)
		} else {
			return fmt.Sprintf("%.3gs", sec)
		}
	}

	var buffer bytes.Buffer
	writers := []io.Writer{&buffer}
	if len(config.Writers) > 0 {
		writers = append(writers, config.Writers...)
	}
	multiWriter := io.MultiWriter(writers...)

	start := time.Now()
	if !config.DisableLog {
		log.Info().Str("cmd", cmd.String()).CallerSkipFrame(1).Send()
	}

	cmd.Stdout = multiWriter
	cmd.Stderr = multiWriter

	err := cmd.Run()
	output := buffer.String()

	if !config.DisableLog {
		log.Info().Str("cmd", cmd.String()).Str("output", output).Err(err).Str("duration", formatDuration(time.Since(start))).CallerSkipFrame(1).Send()
	}

	return output, err
}
