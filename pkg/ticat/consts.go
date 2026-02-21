package ticat

import (
	"os"
	"path/filepath"
)

var (
	SelfName    string = getSelfName()
	ModsRepoExt string = "." + SelfName
	MetaExt     string = "." + SelfName
)

func getSelfName() string {
	execPath, err := os.Executable()
	if err != nil {
		return "ticat"
	}
	return filepath.Base(execPath)
}

const (
	ListSep                  string = ","
	CmdRootDisplayName       string = "<root>"
	CmdBuiltinName           string = "builtin"
	CmdBuiltinDisplayName    string = "<builtin>"
	Spaces                   string = "\t\n\r "
	AbbrsSep                 string = "|"
	EnvOpSep                 string = ":"
	SequenceSep              string = ":"
	CmdPathSep               string = "."
	FakeCmdPathSepSuffixs    string = "/\\"
	CmdPathAlterSeps         string = "."
	EnvBracketLeft           string = "{"
	EnvBracketRight          string = "}"
	EnvKeyValSep             string = "="
	EnvPathSep               string = "."
	SysArgPrefix             string = "%"
	EnvValDelAllMark         string = "--"
	EnvRuntimeSysPrefix      string = "sys"
	EnvStrsPrefix            string = "strs"
	EnvFileName              string = "bootstrap.env"
	ProtoSep                 string = "\t"
	FlowExt                  string = ".tiflow"
	HelpExt                  string = ".tihelp"
	HubFileName              string = "repos.hub"
	ReposFileName            string = "hub.ticat"
	SessionEnvFileName       string = "env"
	SessionStatusFileName    string = "status"
	FlowTemplateBracketLeft  string = "[["
	FlowTemplateBracketRight string = "]]"
	FlowTemplateMultiplyMark string = "*"
	TagMark                  string = "@"
	TrivialMark              string = "^"
	TagOutOfTheBox           string = TagMark + "ready"
	TagProvider              string = TagMark + "config"
	TagSelfTest              string = TagMark + "selftest"
	EnvSnapshotExt           string = ".env"
	CmdPathSession           string = "sessions"
	ArgEnumSep               string = "|"
)
