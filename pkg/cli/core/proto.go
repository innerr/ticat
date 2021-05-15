package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func EnvOutput(env *Env, writer io.Writer, protoEnvMark string, protoSep string) error {
	if env.Parent() != nil {
		EnvOutput(env.Parent(), writer, protoEnvMark, protoSep)
	}

	protoPrefix := ""
	if len(protoEnvMark) != 0 {
		protoPrefix = protoEnvMark + protoSep
	}

	keys, vals := env.Pairs()
	for i, k := range keys {
		v := vals[i]
		// TODO: "strs.proto-sep" can't be save, handle it better
		if len(strings.TrimSpace(v.Raw)) == 0 {
			continue
		}
		_, err := fmt.Fprintf(writer, "%s%s%s%s%s%s\n", protoPrefix,
			k, protoSep, v.Raw, protoSep, env.LayerType())
		if err != nil {
			return err
		}
	}
	return nil
}

func EnvInput(env *Env, reader io.Reader, protoEnvMark string, protoSep string) (rest []string, err error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := strings.Trim(scanner.Text(), "\n\r")
		if !strings.HasPrefix(text, protoEnvMark) {
			rest = append(rest, text)
			continue
		}
		fields := strings.Split(text, protoSep)
		if len(fields) != 3 && len(fields) != 4 {
			rest = append(rest, text)
			continue
		}
		key := fields[1]
		val := fields[2]
		env.Set(key, val)
	}

	return rest, nil
}

func SaveEnvToFile(env *Env, path string, protoEnvMark string, protoSep string) {
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(fmt.Errorf("[SaveEnvToFile] open env file '%s' failed: %v", tmp, err))
	}
	defer file.Close()

	err = EnvOutput(env, file, protoEnvMark, protoSep)
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
