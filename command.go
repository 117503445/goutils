package goutils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var CommandLogger = Logger.With().Str("module", "goutils.command").Logger()

// Deprecated: Use Exec instead
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
	CommandLogger.Debug().Str("cwd", cwd).Str("command", commandStr).Msg("Run Command")
	cmd := exec.Command(command, args...)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	err = cmd.Run()
	if err != nil {
		CommandLogger.Error().Err(err).Str("cwd", cwd).Str("command", command).Strs("args", args).
			Msg("Failed to run command")
		return fmt.Errorf("failed to run command: %w", err)
	}
	CommandLogger.Debug().Str("cwd", cwd).Str("command", commandStr).Msg("Run Command Done")

	return nil
}

type ExecOptions struct {
	// Cwd is the working directory of the command. If empty, the current working directory is used.
	Cwd string

	PreExecHandler  func(*ExecOptions)
	ExecutedHandler func(*ExecOptions, *ExecResult, error)
}

// preExecHandlerLog is the default pre-execution handler
var preExecHandlerLog = func(o *ExecOptions) {
	CommandLogger.Debug().Str("cwd", o.Cwd).Msg("Run Command")
}

var executedHandlerErrorLog = func(o *ExecOptions, r *ExecResult, err error) {
	if err != nil {
		CommandLogger.Error().Err(err).Str("cwd", o.Cwd).Str("command", r.Output).Msg("Failed to run command")
	}
}
var executedHandlerFatalLog = func(o *ExecOptions, r *ExecResult, err error) {
	if err != nil {
		CommandLogger.Fatal().Err(err).Str("cwd", o.Cwd).Str("command", r.Output).Msg("Failed to run command")
	}
}

// ExecOpt is the default options for Exec
var ExecOpt = &ExecOptions{
	Cwd:             "",
	PreExecHandler:  preExecHandlerLog,
	ExecutedHandler: executedHandlerFatalLog,
}

type execOption interface {
	applyTo(*ExecOptions) error
}

type WithCwd string

func (w WithCwd) applyTo(o *ExecOptions) error {
	o.Cwd = string(w)
	return nil
}

type WithWorkDirCmd struct {
}

func (w WithWorkDirCmd) applyTo(o *ExecOptions) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	o.Cwd = wd
	return nil
}

type WithPreExecLog struct {
}

func (w WithPreExecLog) applyTo(o *ExecOptions) error {
	o.PreExecHandler = preExecHandlerLog
	return nil
}

type WithPreExecSlient struct {
}

func (w WithPreExecSlient) applyTo(o *ExecOptions) error {
	o.PreExecHandler = func(o *ExecOptions) {}
	return nil
}

type WithExecutedHandlerErrorLog struct {
}

func (w WithExecutedHandlerErrorLog) applyTo(o *ExecOptions) error {
	o.ExecutedHandler = executedHandlerErrorLog
	return nil
}

type WithExecutedHandlerFatalLog struct {
}

func (w WithExecutedHandlerFatalLog) applyTo(o *ExecOptions) error {
	o.ExecutedHandler = executedHandlerFatalLog
	return nil
}

type WithExecutedHandlerSlient struct {
}

func (w WithExecutedHandlerSlient) applyTo(o *ExecOptions) error {
	o.ExecutedHandler = func(o *ExecOptions, r *ExecResult, err error) {}
	return nil
}

// WithExeParentDir is a option to set the working directory to the parent directory of the executable
type WithExeParentDir struct {
}

func (w WithExeParentDir) applyTo(o *ExecOptions) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exe)
	o.Cwd = exeDir
	return nil
}

// ExecResult is the result of the command
type ExecResult struct {
	Stdout string
	Stderr string

	// Output is the combined stdout and stderr
	Output string
}

type resultWriter struct {
	isStdout bool
	isStderr bool
	result   *ExecResult
}

func (w *resultWriter) Write(p []byte) (n int, err error) {
	if w.isStdout {
		w.result.Stdout += string(p)
	}
	if w.isStderr {
		w.result.Stderr += string(p)
	}
	w.result.Output += string(p)
	return len(p), nil
}

// Exec is a wrapper of exec.Command.
//
// Parameters:
// - cmd: the command to run, e.g. "ls -l". Spaces are used to split the command and arguments. Shell features like pipes are not supported.
// - opts: options to customize the behavior of the command
//
// Returns:
// - *ExecResult: the result of the command. Always not nil. Even if the command fails, the result may contain some output.
// - error: if the command fails
func Exec(cmd string, opts ...execOption) (*ExecResult, error) {
	r := &ExecResult{}

	opt := ExecOpt
	for _, o := range opts {
		err := o.applyTo(opt)
		if err != nil {
			return r, err
		}
	}

	strs := strings.Split(cmd, " ")
	if len(strs) == 0 {
		return r, fmt.Errorf("empty command")
	}
	name := strs[0]

	command := exec.Command(name, strs[1:]...)
	command.Dir = opt.Cwd
	command.Stdout = &resultWriter{isStdout: true, result: r}
	command.Stderr = &resultWriter{isStderr: true, result: r}

	if opt.PreExecHandler != nil {
		opt.PreExecHandler(opt)
	}

	err := command.Run()
	if opt.ExecutedHandler != nil {
		opt.ExecutedHandler(opt, r, err)
	}
	
	return r, err
}
