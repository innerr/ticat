package cli

import (
	"fmt"
	"os"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func GenEnvFromStdin(protoEnvMark string, protoSep string) *core.Env {
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(fmt.Errorf("[GenEnvFromStdin] get stdin stat failed %v", err))
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil
	}
	env := core.NewEnv()
	rest, err := core.EnvInput(env, os.Stdin, protoEnvMark, protoSep)
	if err != nil {
		panic(fmt.Errorf("[GenEnvFromStdin] parse stdin failed %v", err))
	}
	if len(rest) != 0 {
		panic(fmt.Errorf("[GenEnvFromStdin] lines cant' be parsed '%v'", rest))
	}
	return env
}
