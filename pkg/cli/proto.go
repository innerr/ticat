package cli

import (
	"strings"
	"bufio"
	"io"
	"fmt"
)

func EnvOutput(env *Env, writer io.Writer) error {
	if env.parent != nil {
		EnvOutput(env.parent, writer)
	}
	for k, v := range env.pairs {
		_, err := fmt.Fprintf(writer, "%s%s%s%s%s%s%s\n", ProtoEnvMark,
			ProtoSep, k, ProtoSep, v.Raw, ProtoSep, env.tp)
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

	return nil, nil
}
