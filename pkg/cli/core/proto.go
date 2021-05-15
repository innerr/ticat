package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

func EnvOutput(env *Env, writer io.Writer, sep string) error {
	// TODO: config?
	filtered := []string{
		"session",
		"strs.",
	}
	flatten := env.Flatten(true, filtered, false)
	var keys []string
	for k, _ := range flatten {
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

func EnvInput(env *Env, reader io.Reader, sep string) error {
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
		env.Set(key, val)
	}

	return nil
}

func SaveEnvToFile(env *Env, path string, sep string) {
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(fmt.Errorf("[SaveEnvToFile] open env file '%s' failed: %v", tmp, err))
	}
	defer file.Close()

	err = EnvOutput(env, file, sep)
	if err != nil {
		panic(fmt.Errorf("[SaveEnvToLocal] write env file '%s' failed: %v", tmp, err))
	}
	file.Close()

	err = os.Rename(tmp, path)
	if err != nil {
		panic(fmt.Errorf("[SaveEnvToLocal] rename env file '%s' to '%s' failed: %v",
			tmp, path, err))
	}
}

func LoadEnvFromFile(env *Env, path string, sep string) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(fmt.Errorf("[LoadEnvFromFile] open local env file '%s' failed: %v",
			path, err))
	}
	defer file.Close()

	err = EnvInput(env, file, sep)
	if err != nil {
		panic(fmt.Errorf("[LoadEnvFromFile] read local env file '%s' failed: %v",
			path, err))
	}
}
