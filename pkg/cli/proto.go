package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func EnvOutput(env *Env, writer io.Writer) error {
	if env.Parent() != nil {
		EnvOutput(env.Parent(), writer)
	}
	keys, vals := env.Pairs()
	for i, k := range keys {
		v := vals[i]
		_, err := fmt.Fprintf(writer, "%s%s%s%s%s%s%s\n", ProtoEnvMark,
			ProtoSep, k, ProtoSep, v.Raw, ProtoSep, env.LayerType())
		if err != nil {
			return err
		}
	}
	return nil
}

func EnvInput(env *Env, reader io.Reader) (rest []string, err error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := strings.Trim(scanner.Text(), "\n\r")
		if !strings.HasPrefix(text, ProtoEnvMark) {
			rest = append(rest, text)
			continue
		}
		fields := strings.Split(text, ProtoSep)
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

func GenEnvFromStdin() *Env {
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(fmt.Errorf("[GenEnvFromStdin] get stdin stat failed %v", err))
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil
	}
	env := NewEnv()
	rest, err := EnvInput(env, os.Stdin)
	if err != nil {
		panic(fmt.Errorf("[GenEnvFromStdin] parse stdin failed %v", err))
	}
	if len(rest) != 0 {
		panic(fmt.Errorf("[GenEnvFromStdin] lines cant' be parsed '%v'", rest))
	}
	return env
}
