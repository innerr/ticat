package cli

import (
	"fmt"
	"os"
	"os/exec"
)

type CmdType string

const (
	CmdTypeNormal CmdType = "normal"
	CmdTypePower  CmdType = "power"
	CmdTypeBash   CmdType = "bash"
)

type NormalCmd func(argv ArgVals, cli *Cli, env *Env) (succeeded bool)
type PowerCmd func(argv ArgVals, cli *Cli, env *Env, cmds []ParsedCmd,
	currCmdIdx int) (newCmds []ParsedCmd, newCurrCmdIdx int, succeeded bool)

type Cmd struct {
	owner  *CmdTree
	ty     CmdType
	quiet  bool
	args   Args
	normal NormalCmd
	power  PowerCmd
	bash   string
}

func NewCmd(owner *CmdTree, cmd NormalCmd, quiet bool) *Cmd {
	return &Cmd{owner, CmdTypeNormal, quiet, newArgs(), cmd, nil, ""}
}

func NewPowerCmd(owner *CmdTree, cmd PowerCmd, quiet bool) *Cmd {
	return &Cmd{owner, CmdTypePower, quiet, newArgs(), nil, cmd, ""}
}

func NewBashCmd(owner *CmdTree, cmd string) *Cmd {
	return &Cmd{owner, CmdTypeBash, false, newArgs(), nil, nil, cmd}
}

func (self *Cmd) Execute(argv ArgVals, cli *Cli, env *Env, cmds []ParsedCmd, currCmdIdx int) ([]ParsedCmd, int, bool) {
	switch self.ty {
	case CmdTypePower:
		return self.power(argv, cli, env, cmds, currCmdIdx)
	case CmdTypeNormal:
		return cmds, currCmdIdx, self.normal(argv, cli, env)
	case CmdTypeBash:
		return cmds, currCmdIdx, self.executeBash(argv, cli, env)
	default:
		panic(fmt.Errorf("[Cmd.Execute] unknown cmd executable type: %v", self.ty))
	}
}

func (self *Cmd) IsPowerCmd() bool {
	return self.ty == CmdTypePower
}

func (self *Cmd) AddArg(name string, defVal string, abbrs ...string) *Cmd {
	self.args.AddArg(self.owner, name, defVal, abbrs...)
	return self
}

func (self *Cmd) executeBash(argv ArgVals, cli *Cli, env *Env) bool {
	var args []string
	args = append(args, self.bash)
	for _, k := range self.args.List() {
		args = append(args, argv[k].Raw)
	}
	cmd := exec.Command("bash", args...)

	errPrefix := "[ERR] execute bash fail: %v"

	osStdout := os.Stdout
	cmd.Stdout = os.Stdout
	defer func() {
		os.Stdout = osStdout
	}()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cli.Screen.Println(fmt.Sprintf(errPrefix, err))
		return false
	}
	defer stdin.Close()

	// Input to bash
	go func() {
		defer stdin.Close()
		EnvOutput(env, stdin)
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cli.Screen.Println(fmt.Sprintf(errPrefix, err))
		return false
	}
	defer stderr.Close()

	err = cmd.Start()
	if err != nil {
		cli.Screen.Println(fmt.Sprintf(errPrefix, err))
		cli.Screen.Println("  - path: " + self.bash)
		for i, arg := range args[1:] {
			cli.Screen.Println(fmt.Sprintf("  - arg:%d %s", i, arg))
		}
		return false
	}

	// Output from bash
	stderrLines, err := EnvInput(env.GetLayer(EnvLayerSession), stderr)

	err = cmd.Wait()
	if err != nil {
		cli.Screen.Println(fmt.Sprintf(errPrefix, err))
	}

	if len(stderrLines) != 0 {
		cli.Screen.Println("\n[stderr]")
		for _, line := range stderrLines {
			cli.Screen.Println("    " + line)
		}
	}

	return err == nil
}
