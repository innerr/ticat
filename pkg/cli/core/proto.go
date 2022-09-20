package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func EnvOutput(env *Env, writer io.Writer, sep string, filtered []string, skipDefault bool) error {
	defEnv := env.GetLayer(EnvLayerDefault)

	flatten := env.Flatten(true, filtered, false)
	var keys []string
	for k, v := range flatten {
		if skipDefault && defEnv.GetRaw(k) == v {
			continue
		}
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		v := env.GetRaw(k)
		_, err := fmt.Fprintf(writer, "%s%s%s\n", k, sep, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func EnvInput(env *Env, reader io.Reader, sep string, delMark string) error {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasSuffix(text, "\n") {
			text = text[0 : len(text)-1]
		}
		i := strings.Index(text, sep)
		if i < 0 {
			return fmt.Errorf("[EnvInput] bad format line '%s', sep '%s'",
				text, sep)
		}
		key := text[0:i]
		val := text[i+1:]
		if val == delMark {
			env.Delete(key)
		} else {
			env.Set(key, val)
		}
	}

	return nil
}

func saveEnvToFile(env *Env, path string, sep string, filtered []string, skipDefault bool) {
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(fmt.Errorf("[SaveEnvToFile] open env file '%s' failed: %v", tmp, err))
	}
	defer file.Close()

	err = EnvOutput(env, file, sep, filtered, skipDefault)
	if err != nil {
		panic(fmt.Errorf("[SaveEnvToFile] write env file '%s' failed: %v", tmp, err))
	}
	file.Close()

	err = os.Rename(tmp, path)
	if err != nil {
		panic(fmt.Errorf("[SaveEnvToFile] rename env file '%s' to '%s' failed: %v",
			tmp, path, err))
	}
}

func SaveEnvToFile(env *Env, path string, sep string, skipDefault bool) {
	// TODO: move to default config
	filtered := []string{
		"session",
		"strs.",
		"display.height",
		"display.width.max",
		"display.executor.displayed",
		"sys.stack",
		"sys.stack-depth",
		"sys.session.",
		"sys.interact",
		"sys.step-by-step",
		"sys.breakpoint",
		"sys.execute-wait-sec",
		"sys.event.",
		"sys.paths.",
	}
	saveEnvToFile(env, path, sep, filtered, skipDefault)
}

func LoadEnvFromFile(env *Env, path string, sep string, delMark string) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(fmt.Errorf("[LoadEnvFromFile] open local env file '%s' failed: %v",
			path, err))
	}
	defer file.Close()

	err = EnvInput(env, file, sep, delMark)
	if err != nil {
		panic(fmt.Errorf("[LoadEnvFromFile] read local env file '%s' failed: %v",
			path, err))
	}
}

func saveEnvToSessionFile(cc *Cli, env *Env, parsedCmd ParsedCmd, skipDefault bool) (sessionDir string, sessionPath string) {
	sep := cc.Cmds.Strs.EnvKeyValSep

	sessionDir = env.GetRaw("session")
	if len(sessionDir) == 0 {
		panic(NewCmdError(parsedCmd, "[Cmd.executeFile] session dir not found in env"))
	}
	sessionFileName := env.GetRaw("strs.session-env-file")
	if len(sessionFileName) == 0 {
		panic(NewCmdError(parsedCmd, "[Cmd.executeFile] session env file name not found in env"))
	}
	sessionPath = filepath.Join(sessionDir, sessionFileName)

	filtered := []string{
		//"display.height",
		//"sys.stack",
		//"sys.stack-depth",
		//"sys.session.",
		//"sys.interact",
	}
	saveEnvToFile(env.GetLayer(EnvLayerSession), sessionPath, sep, filtered, skipDefault)
	return
}
