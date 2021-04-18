package builtin

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pingcap/ticat/pkg/cli"
)

func Sleep(_ *cli.Cli, env *cli.Env) bool {
	durStr := env.Get("sleep.duration").Raw
	// Default unit is 's'
	_, err := strconv.ParseFloat(durStr, 64)
	if err == nil {
		durStr += "s"
	}

	dur, err := time.ParseDuration(durStr)
	if err != nil {
		fmt.Printf("[ERR] %v\n", err)
		return false
	}
	time.Sleep(dur)
	return true
}

func Dummy(_ *cli.Cli, env *cli.Env) bool {
	fmt.Println("Dummy cmd here")
	return true
}
