package gexec_test

import (
	"io"
	"os"
	"testing"

	"github.com/117503445/goutils"
	"github.com/117503445/goutils/gexec"
)

func TestCMD(t *testing.T) {
	goutils.InitZeroLog(goutils.WithNoColor{})

	cmd := gexec.Commands([]string{"ls", "-l"})
	cmd.Run()

	// gexec.Run(gexec.Command("echo 1"))

	// gexec.Run(gexec.Commands([]string{"bash", "-c", "echo $SHELL"}))

	// cmd = gexec.Commands([]string{"bash", "-c", "echo $A"})
	// gexec.SetEnvs(cmd, map[string]string{
	// 	"A": "1",
	// })
	// gexec.Run(cmd)

	gexec.Run(
		gexec.SetPwd(
			"/tmp",
			gexec.SetEnvs(
				map[string]string{
					"A": "1",
				},
				gexec.Commands(
					[]string{"bash", "-c", "pwd && echo $A && exit -4"},
				),
			),
		),
		&gexec.RunCfg{
			DisableLog: false,
			Writers: []io.Writer{
				os.Stdout,
			},
		},
	)

}
