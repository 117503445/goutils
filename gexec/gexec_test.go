package gexec_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/117503445/goutils"
	"github.com/117503445/goutils/gexec"
)

func TestCMD(t *testing.T) {
	ast := assert.New(t)
	goutils.InitZeroLog(goutils.WithNoColor{})

	// Create temporary directory for this test
	tempDir := t.TempDir()

	// Test basic command execution
	cmd := gexec.Commands([]string{"echo", "hello"})
	err := cmd.Run()
	ast.NoError(err)

	// Test Run function with environment variables
	_, err = gexec.Run(
		gexec.SetPwd(
			tempDir,
			gexec.SetEnvs(
				map[string]string{
					"A": "1",
				},
				gexec.Commands(
					[]string{"bash", "-c", "pwd && echo $A"},
				),
			),
		),
		&gexec.RunCfg{
			DisableLog: true, // Disable logging for cleaner test output
		},
	)
	ast.NoError(err)
}
