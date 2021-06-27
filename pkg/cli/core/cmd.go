package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mattn/go-shellwords"
)

type CmdType string

const (
	CmdTypeUninited   CmdType = "uninited"
	CmdTypeNormal     CmdType = "normal"
	CmdTypePower      CmdType = "power"
	CmdTypeFlow       CmdType = "flow"
	CmdTypeEmpty      CmdType = "no-executable"
	CmdTypeFile       CmdType = "executable-file"
	CmdTypeEmptyDir   CmdType = "dir-with-no-executable"
	CmdTypeDirWithCmd CmdType = "dir-with-executable-file"
)

type NormalCmd func(argv ArgVals, cc *Cli, env *Env, cmd ParsedCmd) (succeeded bool)
type PowerCmd func(argv ArgVals, cc *Cli, env *Env, flow *ParsedCmds,
	currCmdIdx int) (newCurrCmdIdx int, succeeded bool)

type Cmd struct {
	owner        *CmdTree
	help         string
	ty           CmdType
	quiet        bool
	priority     bool
	args         Args
	normal       NormalCmd
	power        PowerCmd
	cmdLine      string
	flow         []string
	envOps       EnvOps
	source       string
	depends      []Depend
	metaFilePath string
	val2env      *Val2Env
	arg2env      *Arg2Env
}

func defaultCmd(owner *CmdTree, help string) *Cmd {
	return &Cmd{
		owner:        owner,
		help:         help,
		ty:           CmdTypeUninited,
		quiet:        false,
		priority:     false,
		args:         newArgs(),
		normal:       nil,
		power:        nil,
		cmdLine:      "",
		flow:         nil,
		envOps:       newEnvOps(),
		source:       "",
		depends:      nil,
		metaFilePath: "",
		val2env:      newVal2Env(),
		arg2env:      newArg2Env(),
	}
}

func NewCmd(owner *CmdTree, help string, cmd NormalCmd) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeNormal
	c.normal = cmd
	return c
}

func NewEmptyCmd(owner *CmdTree, help string) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeEmpty
	return c
}

func NewPowerCmd(owner *CmdTree, help string, cmd PowerCmd) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypePower
	c.power = cmd
	return c
}

func NewFileCmd(owner *CmdTree, help string, cmd string) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeFile
	c.cmdLine = cmd
	return c
}

func NewEmptyDirCmd(owner *CmdTree, help string, dir string) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeEmptyDir
	c.cmdLine = dir
	return c
}

func NewDirWithCmd(owner *CmdTree, help string, cmd string) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeDirWithCmd
	c.cmdLine = cmd
	return c
}

func NewFlowCmd(owner *CmdTree, help string, flow []string) *Cmd {
	c := defaultCmd(owner, help)
	c.ty = CmdTypeFlow
	c.flow = flow
	return c
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
		return currCmdIdx, self.normal(argv, cc, env, flow.Cmds[currCmdIdx])
	case CmdTypeFile:
		return currCmdIdx, self.executeFile(argv, cc, env)
	case CmdTypeEmptyDir:
		return currCmdIdx, true
	case CmdTypeDirWithCmd:
		return currCmdIdx, self.executeFile(argv, cc, env)
	case CmdTypeFlow:
		return currCmdIdx, self.executeFlow(argv, cc, env)
	case CmdTypeEmpty:
		return currCmdIdx, true
	default:
		panic(fmt.Errorf("[Cmd.Execute] unknown cmd executable type: %v", self.ty))
	}
}

func (self *Cmd) MatchFind(findStr string) bool {
	if strings.Index(self.owner.DisplayPath(), findStr) >= 0 {
		return true
	}
	if strings.Index(self.help, findStr) >= 0 {
		return true
	}
	if strings.Index(self.cmdLine, findStr) >= 0 {
		return true
	}
	if self.args.MatchFind(findStr) {
		return true
	}
	if self.val2env.MatchFind(findStr) {
		return true
	}
	if self.arg2env.MatchFind(findStr) {
		return true
	}
	if self.envOps.MatchFind(findStr) {
		return true
	}
	if strings.Index(string(self.ty), findStr) >= 0 {
		return true
	}
	for _, dep := range self.depends {
		if strings.Index(dep.OsCmd, findStr) >= 0 {
			return true
		}
		if strings.Index(dep.Reason, findStr) >= 0 {
			return true
		}
	}
	if len(self.source) == 0 {
		if strings.Index("builtin", findStr) >= 0 {
			return true
		}
	} else {
		if strings.Index(self.source, findStr) >= 0 {
			return true
		}
	}
	if self.quiet && strings.Index("quiet", findStr) >= 0 {
		return true
	}
	if self.ty == CmdTypePower && strings.Index("power", findStr) >= 0 {
		return true
	}
	if self.priority && strings.Index("priority", findStr) >= 0 {
		return true
	}
	return false
}

func (self *Cmd) AddArg(name string, defVal string, abbrs ...string) *Cmd {
	self.args.AddArg(self.owner, name, defVal, abbrs...)
	return self
}

func (self *Cmd) AddEnvOp(name string, op uint) *Cmd {
	self.envOps.AddOp(name, op)
	return self
}

func (self *Cmd) AddSub(name string, abbrs ...string) *CmdTree {
	return self.owner.AddSub(name, abbrs...)
}

func (self *Cmd) SetSource(s string) *Cmd {
	self.source = s
	return self
}

func (self *Cmd) SetMetaFile(path string) *Cmd {
	self.metaFilePath = path
	return self
}

func (self *Cmd) AddDepend(dep string, reason string) *Cmd {
	self.depends = append(self.depends, Depend{dep, reason})
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

func (self *Cmd) AddVal2Env(envKey string, val string) *Cmd {
	self.val2env.Add(envKey, val)
	return self
}

func (self *Cmd) AddArg2Env(envKey string, argName string) *Cmd {
	self.arg2env.Add(envKey, argName)
	return self
}

func (self *Cmd) GetVal2Env() *Val2Env {
	return self.val2env
}

func (self *Cmd) GetArg2Env() *Arg2Env {
	return self.arg2env
}

func (self *Cmd) GetDepends() []Depend {
	return self.depends
}

func (self *Cmd) MetaFile() string {
	return self.metaFilePath
}

func (self *Cmd) Owner() *CmdTree {
	return self.owner
}

func (self *Cmd) Source() string {
	return self.source
}

func (self *Cmd) Help() string {
	return self.help
}

func (self *Cmd) DisplayHelpStr() string {
	if len(self.help) == 0 && self.ty == CmdTypeFlow {
		return self.cmdLine
	}
	return self.help
}

func (self *Cmd) IsNoExecutableCmd() bool {
	if len(self.val2env.EnvKeys()) > 0 {
		return false
	}
	if len(self.arg2env.EnvKeys()) > 0 {
		return false
	}
	return self.ty == CmdTypeUninited || self.ty == CmdTypeEmpty || self.ty == CmdTypeEmptyDir
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

func (self *Cmd) FlowStrs() []string {
	return self.flow
}

// TODO: move to parser ?
func (self *Cmd) RenderedFlowStrs(
	env *Env,
	allowFlowTemplateRenderError bool) (flow []string, fullyRendered bool) {

	templBracketLeft := self.owner.Strs.FlowTemplateBracketLeft
	templBracketRight := self.owner.Strs.FlowTemplateBracketRight
	hasError := false
	for _, it := range self.flow {
		findPos := 0
		for {
			str := it[findPos:]
			i := strings.Index(str, templBracketLeft)
			if i < 0 {
				break
			}
			tail := str[i+len(templBracketLeft):]
			j := strings.Index(tail, templBracketRight)
			if j < 0 {
				break
			}
			key := tail[0:j]
			// TODO: remove this, not allow env is nil
			if env == nil {
				return self.flow, false
			}
			val, ok := env.GetEx(key)
			if !ok {
				if allowFlowTemplateRenderError {
					hasError = true
					findPos += j+len(templBracketRight)
					continue
				}
				err := CmdMissedEnvValWhenRenderFlow{
					"render flow template failed, env value missed.",
					self.owner.DisplayPath(),
					self.metaFilePath,
					self.source,
					key,
				}
				panic(err)
			}
			it = it[:findPos] + str[0:i] + val.Raw + tail[j+len(templBracketRight):]
		}
		flow = append(flow, it)
	}
	fullyRendered = !hasError
	return
}

func (self *Cmd) Flow(env *Env, allowFlowTemplateRenderError bool) (flow []string, rendered bool) {
	flow, rendered = self.RenderedFlowStrs(env, allowFlowTemplateRenderError)
	if !rendered || len(flow) == 0 {
		return
	}
	flowStr := strings.Join(flow, " ")
	flow, err := shellwords.Parse(flowStr)
	if err != nil {
		// TODO: better display
		panic(fmt.Errorf("[Cmd.executeFlow] parse '%s' failed: %v",
			self.cmdLine, err))
	}
	return
}

func (self *Cmd) IsTheSameFunc(fun interface{}) bool {
	fr1 := reflect.ValueOf(fun)
	if self.power != nil {
		fr2 := reflect.ValueOf(self.power)
		if fr1.Pointer() == fr2.Pointer() {
			return true
		}
	}
	if self.normal != nil {
		fr2 := reflect.ValueOf(self.normal)
		if fr1.Pointer() == fr2.Pointer() {
			return true
		}
	}
	return false
}

func (self *Cmd) executeFlow(argv ArgVals, cc *Cli, env *Env) bool {
	flow, _ := self.Flow(env, false)
	return cc.Executor.Execute(cc, flow...)
}

func (self *Cmd) executeFile(argv ArgVals, cc *Cli, env *Env) bool {
	if len(self.cmdLine) == 0 {
		return true
	}

	for _, dep := range self.depends {
		_, err := exec.LookPath(dep.OsCmd)
		if err != nil {
			// TODO: better display
			panic(fmt.Errorf("[Cmd.executeFile] %s", err))
		}
	}

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

	sep := cc.Cmds.Strs.ProtoSep

	sessionDir, sessionPath := saveEnvToSessionFile(cc, env)

	args = append(args, self.cmdLine)
	args = append(args, sessionDir)
	for _, k := range self.args.Names() {
		args = append(args, argv[k].Raw)
	}
	cmd := exec.Command(bin, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		indent1 := strings.Repeat(" ", 4)
		indent2 := strings.Repeat(" ", 8)
		cc.Screen.Print(fmt.Sprintf("\n[%s] failed:\n", self.owner.DisplayPath()))
		if len(self.args.Names()) != 0 {
			cc.Screen.Print(fmt.Sprintf("%s- args:\n", indent1))
			for _, k := range self.args.Names() {
				cc.Screen.Print(fmt.Sprintf("%s%s = %s\n", indent2,
					strings.Join(self.args.Abbrs(k), self.owner.Strs.AbbrsSep), mayQuoteStr(argv[k].Raw)))
			}
		}
		cc.Screen.Print(fmt.Sprintf("%s- bin:  %s\n", indent1, bin))
		cc.Screen.Print(fmt.Sprintf("%s- file: %s\n", indent1, self.cmdLine))
		cc.Screen.Print(fmt.Sprintf("%s- env:  %s\n", indent1, sessionPath))
		cc.Screen.Print(fmt.Sprintf("%s- err:  %s\n", indent1, err))
		return false
	}

	LoadEnvFromFile(env.GetLayer(EnvLayerSession), sessionPath, sep)
	return true
}

func saveEnvToSessionFile(cc *Cli, env *Env) (sessionDir string, sessionPath string) {
	sep := cc.Cmds.Strs.ProtoSep

	sessionDir = env.GetRaw("session")
	if len(sessionDir) == 0 {
		panic(fmt.Errorf("[Cmd.executeFile] session dir not found in env"))
	}
	sessionFileName := env.GetRaw("strs.session-env-file")
	if len(sessionFileName) == 0 {
		panic(fmt.Errorf("[Cmd.executeFile] session env file name not found in env"))
	}
	sessionPath = filepath.Join(sessionDir, sessionFileName)
	SaveEnvToFile(env.GetLayer(EnvLayerSession), sessionPath, sep)
	return
}

func mayQuoteStr(origin string) string {
	trimed := strings.TrimSpace(origin)
	if len(trimed) == 0 || len(trimed) != len(origin) {
		return "'" + origin + "'"
	}
	fields := strings.Fields(origin)
	if len(fields) != 1 {
		return "'" + origin + "'"
	}
	return origin
}

type Depend struct {
	OsCmd  string
	Reason string
}

type CmdError struct {
	Cmd ParsedCmd
	Err error
}

func WrapCmdError(cmd ParsedCmd, err error) *CmdError {
	return &CmdError{cmd, err}
}

func NewCmdError(cmd ParsedCmd, err string) *CmdError {
	return &CmdError{cmd, fmt.Errorf(err)}
}

func (self CmdError) Error() string {
	return self.Err.Error()
}
