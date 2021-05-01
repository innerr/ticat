package cli

const (
	CmdPathSep       string = "."
	CmdPathAlterSeps string = "./"
	Spaces           string = "\t\n\r "
	EnvBracketLeft   string = "{"
	EnvBracketRight  string = "}"
	EnvKeyValSep     string = "="
	EnvPathSep       string = "."
	EnvValDelMark    string = "-"
	EnvValDelAllMark string = "--"
	AbbrSep          string = "|"

	SelfName            string = "ticat"
	EnvRuntimeSysPrefix string = "sys."
	CmdRootDisplayName  string = "<root>"
	ErrStrPrefix        string = "[ERR] "

	ProtoMark    string = "proto." + SelfName
	ProtoEnvMark string = ProtoMark + ".env"
	ProtoSep     string = "\t"
)
