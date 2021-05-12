package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CmdType string

const (
	CmdTypeNormal CmdType = "normal"
	CmdTypePower  CmdType = "power"
	CmdTypeFile   CmdType = "file"
	CmdTypeDir    CmdType = "dir"
	CmdTypeFlow   CmdType = "flow"
)

type NormalCmd func(argv ArgVals, cc *Cli, env *Env) (succeeded bool)
type PowerCmd func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds,
	currCmdIdx int) (newCurrCmdIdx int, succeeded bool)

type Cmd struct {
	owner    *CmdTree
	help     string
	ty       CmdType
	quiet    bool
	priority bool
	args     Args
	normal   NormalCmd
	power    PowerCmd
	cmdLine  string
	envOps   EnvOps
}

func NewCmd(owner *CmdTree, help string, cmd NormalCmd) *Cmd {
	return &Cmd{owner, help, CmdTypeNormal, false, false,
		newArgs(), cmd, nil, "", newEnvOps()}
}

func NewPowerCmd(owner *CmdTree, help string, cmd PowerCmd) *Cmd {
	return &Cmd{owner, help, CmdTypePower, false, false,
		newArgs(), nil, cmd, "", newEnvOps()}
}

func NewFileCmd(owner *CmdTree, help string, cmd string) *Cmd {
	return &Cmd{owner, help, CmdTypeFile, false, false,
		newArgs(), nil, nil, cmd, newEnvOps()}
}

func NewDirCmd(owner *CmdTree, help string, cmd string) *Cmd {
	return &Cmd{owner, help, CmdTypeDir, false, false,
		newArgs(), nil, nil, cmd, newEnvOps()}
}

func NewFlowCmd(owner *CmdTree, help string, flow string) *Cmd {
	return &Cmd{owner, help, CmdTypeFlow, false, false,
		newArgs(), nil, nil, flow, newEnvOps()}
}

func (self *Cmd) Execute(
	argv ArgVals,
	cc *Cli,
	env *Env,
	flow *ParsedCmds,
	currCmdIdx int) (int, bool) {

	switch self.ty {
	case CmdTypePower:
		return self.power(argv, cc, env, flow, currCmdIdx)
	case CmdTypeNormal:
		return currCmdIdx, self.normal(argv, cc, env)
	case CmdTypeFile:
		return currCmdIdx, self.executeFile(argv, cc, env)
	case CmdTypeDir:
		return currCmdIdx, self.executeFile(argv, cc, env)
	case CmdTypeFlow:
		return currCmdIdx, self.executeFlow(argv, cc, env)
	default:
		panic(fmt.Errorf("[Cmd.Execute] unknown cmd executable type: %v", self.ty))
	}
}

func (self *Cmd) AddArg(name string, defVal string, abbrs ...string) *Cmd {
	self.args.AddArg(self.owner, name, defVal, abbrs...)
	return self
}

func (self *Cmd) AddEnvOp(name string, op uint) *Cmd {
	self.envOps.AddOp(name, op)
	return self
}

func (self *Cmd) SetQuiet() *Cmd {
	self.quiet = true
	return self
}

func (self *Cmd) SetPriority() *Cmd {
	self.priority = true
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

func (self *Cmd) IsPriority() bool {
	return self.priority
}

func (self *Cmd) Type() CmdType {
	return self.ty
}

func (self *Cmd) CmdLine() string {
	return self.cmdLine
}

func (self *Cmd) Args() Args {
	return self.args
}

func (self *Cmd) EnvOps() EnvOps {
	return self.envOps
}

func (self *Cmd) executeFlow(argv ArgVals, cc *Cli, env *Env) bool {
	return cc.Executor.Execute(cc, false, strings.Fields(self.cmdLine)...)
}

func (self *Cmd) executeFile(argv ArgVals, cc *Cli, env *Env) bool {
	var bin string
	var args []string
	ext := filepath.Ext(self.cmdLine)

	// TODO: move this code block out?
	runner := env.Get("sys.ext.exec" + ext).Raw
	if len(runner) != 0 {
		fields := strings.Fields(runner)
		if len(fields) == 1 {
			bin = runner
		} else {
			bin = fields[0]
			args = append(args, fields[1:]...)
		}
	} else {
		bin = "bash"
	}

	args = append(args, self.cmdLine)
	for _, k := range self.args.Names() {
		args = append(args, argv[k].Raw)
	}
	cmd := exec.Command(bin, args...)

	errPrefix := "[ERR] execute file failed: %v\n"

	osStdout := os.Stdout
	cmd.Stdout = os.Stdout
	defer func() {
		os.Stdout = osStdout
	}()

	printLine := func(text string) {
		cc.Screen.Print(text + "\n")
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		printLine(fmt.Sprintf(errPrefix, err))
		return false
	}
	defer stdin.Close()

	// Input to bash
	go func() {
		defer stdin.Close()
		EnvOutput(env, stdin, self.owner.Strs.ProtoEnvMark, self.owner.Strs.ProtoSep)
	}()

	// TODO: use named-pipe (mkfifo) instead of stderr?
	stderr, err := cmd.StderrPipe()
	if err != nil {
		printLine(fmt.Sprintf(errPrefix, err))
		return false
	}
	defer stderr.Close()

	// TODO: better stderr output
	err = cmd.Start()
	if err != nil {
		printLine(fmt.Sprintf(errPrefix, err))
		printLine("  - path: " + self.cmdLine)
		for i, arg := range args[1:] {
			printLine(fmt.Sprintf("  - arg:%d %s", i, arg))
		}
		return false
	}

	// The output result from bash command's stderr
	stderrLines, err := EnvInput(env.GetLayer(EnvLayerSession), stderr,
		self.owner.Strs.ProtoEnvMark, self.owner.Strs.ProtoSep)

	err = cmd.Wait()
	if err != nil {
		printLine(fmt.Sprintf(errPrefix, err))
	}

	if len(stderrLines) != 0 {
		printLine("\n[stderr]")
		for _, line := range stderrLines {
			printLine("    " + line)
		}
	}

	return err == nil
}
