package builtin

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func InteractiveMode(cc *core.Cli, env *core.Env, exitStr string) {
	selfName := env.GetRaw("strs.self-name")
	cc.Screen.Print(display.ColorExplain("", env) + display.ColorWarn(exitStr, env) +
		display.ColorExplain(": exit interactive mode\n", env))

	sessionEnv := env.GetLayer(core.EnvLayerSession)
	sessionEnv.SetBool("sys.interact.inside", true)

	cc = cc.CopyForInteract()
	buf := bufio.NewReader(os.Stdin)
	for {
		if env.GetBool("sys.interact.leaving") {
			break
		}
		cc.Screen.Print(display.ColorTip(selfName+"> ", env))
		lineBytes, err := buf.ReadBytes('\n')
		if err != nil {
			panic(fmt.Errorf("[readFromStdin] read from stdin failed: %v", err))
		}
		if len(lineBytes) == 0 {
			continue
		}
		line := strings.TrimSpace(string(lineBytes))
		if line == exitStr {
			break
		}
		executorSafeExecute("(interact)", cc, env, nil, core.FlowStrToStrs(line)...)
	}

	sessionEnv.GetLayer(core.EnvLayerSession).Delete("sys.interact.inside")
}

func executorSafeExecute(caller string, cc *core.Cli, env *core.Env, masks []*core.ExecuteMask, input ...string) {
	env = env.GetLayer(core.EnvLayerSession)
	stackDepth := env.GetRaw("sys.stack-depth")
	stack := env.GetRaw("sys.stack")

	defer func() {
		env.Set("sys.stack-depth", stackDepth)
		env.Set("sys.stack", stack)

		if !env.GetBool("sys.panic.recover") {
			return
		}
		if r := recover(); r != nil {
			display.PrintError(cc, env, r.(error))
		}
	}()
}
