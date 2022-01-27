package builtin

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

const (
	joinKvName = "join.kvs"
)

type joinKv struct {
	Key  string   `json:"key"`
	Vals []string `json:"vals"`
}

func JoinNew(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	key := argv.GetRaw("key")
	val := argv.GetRaw("value")

	if key == "" || val == "" {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx],
			"need key or values"))
	}
	vals := strings.Split(val, env.GetRaw("strs.list-sep"))
	kv := joinKv{
		Key:  key,
		Vals: vals,
	}
	kvsStr := env.GetRaw(joinKvName)
	kvs := []joinKv{}
	if kvsStr != "" {
		if err := json.Unmarshal([]byte(kvsStr), &kvs); err != nil {
			panic(core.NewCmdError(flow.Cmds[currCmdIdx],
				fmt.Sprintf("parse argument '%s' failed: %s", kvsStr, err.Error())))
		}
	}
	kvs = append(kvs, kv)
	data, err := json.Marshal(kvs)
	if err != nil {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("encode argument '%s' failed: %s", kvsStr, err.Error())))
	}
	env.GetLayer(core.EnvLayerSession).Set(joinKvName, string(data))
	return currCmdIdx, true
}

func JoinRun(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	kvs := []joinKv{}
	kvsStr := env.GetRaw(joinKvName)
	if err := json.Unmarshal([]byte(kvsStr), &kvs); err != nil {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("parse argument '%s' failed: %s", kvsStr, err.Error())))
	}
	cmd := argv.GetRaw("cmd")
	if cmd == "" {
		panic(core.NewCmdError(flow.Cmds[currCmdIdx],
			"can't execute null ticat command"))
	}
	iter := newJoinKvsIter(kvs)
	for {
		args := []string{cmd}
		keys, vals := iter.get()
		for i, key := range keys {
			args = append(args, key+"="+vals[i])
		}
		cc.Executor.Execute(flow.Cmds[currCmdIdx].DisplayPath(cc.Cmds.Strs.PathSep, false),
			true, cc, env, nil, args...)
		if !iter.next() {
			break
		}
	}
	return currCmdIdx, true
}

type joinKvsIter struct {
	kvs []joinKv
	ids []int
}

func newJoinKvsIter(kvs []joinKv) *joinKvsIter {
	iter := &joinKvsIter{
		kvs: kvs,
		ids: make([]int, len(kvs)),
	}
	return iter
}

func (iter *joinKvsIter) next() (ok bool) {
	l := len(iter.ids)
	for i := l - 1; i >= 0; i-- {
		if iter.ids[i] < len(iter.kvs[i].Vals)-1 {
			ok = true
			break
		}
	}
	if ok {
		for i := l - 1; i >= 0; i-- {
			if iter.ids[i] < len(iter.kvs[i].Vals)-1 {
				iter.ids[i]++
				break
			} else {
				iter.ids[i] = 0
			}
		}
	}
	return
}

func (iter *joinKvsIter) get() (keys, vals []string) {
	for i := range iter.ids {
		keys = append(keys, iter.kvs[i].Key)
		vals = append(vals, iter.kvs[i].Vals[iter.ids[i]])
	}
	return
}
