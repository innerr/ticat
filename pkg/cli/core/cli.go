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

