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

type NormalCmd func(argv ArgVals, cc *Cli, env *Env) (succeeded bool)
type PowerCmd func(argv ArgVals, cc *Cli, env *Env, cmds []ParsedCmd,
	currCmdIdx int) (newCmds []ParsedCmd, newCurrCmdIdx int, succeeded bool)

type Cmd struct {
	owner  *CmdTree
	help   string
	ty     CmdType
	quiet  bool
	args   Args
	normal NormalCmd
	power  PowerCmd
	bash   string
}

func NewCmd(owner *CmdTree, help string, cmd NormalCmd) *Cmd {
	return &Cmd{owner, help, CmdTypeNormal, false, newArgs(), cmd, nil, ""}
}

func NewPowerCmd(owner *CmdTree, help string, cmd PowerCmd) *Cmd {
	return &Cmd{owner, help, CmdTypePower, false, newArgs(), nil, cmd, ""}
}

func NewBashCmd(owner *CmdTree, help string, cmd string) *Cmd {
	return &Cmd{owner, help, CmdTypeBash, false, newArgs(), nil, nil, cmd}
}

func (self *Cmd) Execute(argv ArgVals, cc *Cli, env *Env, cmds []ParsedCmd, currCmdIdx int) ([]ParsedCmd, int, bool) {
	switch self.ty {
	case CmdTypePower:
		return self.power(argv, cc, env, cmds, currCmdIdx)
	case CmdTypeNormal:
		return cmds, currCmdIdx, self.normal(argv, cc, env)
	case CmdTypeBash:
		return cmds, currCmdIdx, self.executeBash(argv, cc, env)
	default:
		panic(fmt.Errorf("[Cmd.Execute] unknown cmd executable type: %v", self.ty))
	}
}

func (self *Cmd) AddArg(name string, defVal string, abbrs ...string) *Cmd {
	self.args.AddArg(self.owner, name, defVal, abbrs...)
	return self
}

func (self *Cmd) SetQuiet() *Cmd {
	self.quiet = true
	return self
}

func (self *Cmd) Help() string {
	return self.help
}

func (self *Cmd) IsPowerCmd() bool {
	return self.ty == CmdTypePower
}

func (self *Cmd) IsQuiet() bool {
	return self.quiet
}

func (self *Cmd) Type() CmdType {
	return self.ty
}

func (self *Cmd) BashCmdLine() string {
	return self.bash
}

func (self *Cmd) Args() Args {
	return self.args
}

func (self *Cmd) executeBash(argv ArgVals, cc *Cli, env *Env) bool {
	var args []string
	args = append(args, self.bash)
	for _, k := range self.args.Names() {
		args = append(args, argv[k].Raw)
	}
	cmd := exec.Command("bash", args...)

	errPrefix := "[ERR] execute bash failed: %v"

	osStdout := os.Stdout
	cmd.Stdout = os.Stdout
	defer func() {
		os.Stdout = osStdout
	}()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cc.Screen.Println(fmt.Sprintf(errPrefix, err))
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
		cc.Screen.Println(fmt.Sprintf(errPrefix, err))
		return false
	}
	defer stderr.Close()

	err = cmd.Start()
	if err != nil {
		cc.Screen.Println(fmt.Sprintf(errPrefix, err))
		cc.Screen.Println("  - path: " + self.bash)
		for i, arg := range args[1:] {
			cc.Screen.Println(fmt.Sprintf("  - arg:%d %s", i, arg))
		}
		return false
	}

	// Output from bash
	stderrLines, err := EnvInput(env.GetLayer(EnvLayerSession), stderr)

	err = cmd.Wait()
	if err != nil {
		cc.Screen.Println(fmt.Sprintf(errPrefix, err))
	}

	if len(stderrLines) != 0 {
		cc.Screen.Println("\n[stderr]")
		for _, line := range stderrLines {
			cc.Screen.Println("    " + line)
		}
	}

	return err == nil
}
