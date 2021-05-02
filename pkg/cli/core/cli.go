package core

type Screen interface {
	Print(text string)
}

type Cli struct {
	GlobalEnv *Env
	Screen    Screen
	Cmds      *CmdTree
	Parser    CliParser
	EnvAbbrs  *EnvAbbrs
}

func NewCli(env *Env, screen Screen, cmds *CmdTree, parser CliParser) *Cli {
	return &Cli{
		env,
		screen,
		cmds,
		parser,
		NewEnvAbbrs(cmds.Strs.RootDisplayName),
	}
}
