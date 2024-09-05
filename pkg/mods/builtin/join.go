package builtin

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/innerr/ticat/pkg/core/model"
)

const (
	joinKvName = "join.kvs"
)

type joinKv struct {
	Key  string   `json:"key"`
	Vals []string `json:"vals"`
}

func JoinNew(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	key := argv.GetRaw("key")
	val := argv.GetRaw("value")

	if key == "" || val == "" {
		panic(model.NewCmdError(flow.Cmds[currCmdIdx],
			"need key or values"))
	}

	vals := []string{}
	if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") && strings.Count(val, ",") <= 2 && strings.Count(val, ",") >= 1 {
		i := strings.Index(val, ",")
		j := strings.LastIndex(val, ",")
		var startStr, endStr, stepStr string
		if i != j {
			startStr = val[1:i]
			endStr = val[i+1 : j]
			stepStr = val[j+1 : len(val)-1]
		} else {
			startStr = val[1:j]
			endStr = val[j+1 : len(val)-1]
			stepStr = "1"
		}
		if strings.Contains(val, ".") {
			start, _ := strconv.ParseFloat(startStr, 64)
			end, _ := strconv.ParseFloat(endStr, 64)
			step, _ := strconv.ParseFloat(stepStr, 64)
			for ; start <= end; start += step {
				vals = append(vals, fmt.Sprintf("%f", start))
			}
		} else {
			start, _ := strconv.Atoi(startStr)
			end, _ := strconv.Atoi(endStr)
			step, _ := strconv.Atoi(stepStr)
			for ; start <= end; start += step {
				vals = append(vals, strconv.Itoa(start))
			}
		}
	} else {
		vals = strings.Split(val, env.GetRaw("strs.list-sep"))
	}

	kv := joinKv{
		Key:  key,
		Vals: vals,
	}
	kvsStr := env.GetRaw(joinKvName)
	kvs := []joinKv{}
	if kvsStr != "" {
		if err := json.Unmarshal([]byte(kvsStr), &kvs); err != nil {
			panic(model.NewCmdError(flow.Cmds[currCmdIdx],
				fmt.Sprintf("parse argument '%s' failed: %s", kvsStr, err.Error())))
		}
	}
	kvs = append(kvs, kv)
	data, err := json.Marshal(kvs)
	if err != nil {
		panic(model.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("encode argument '%s' failed: %s", kvsStr, err.Error())))
	}
	env.GetLayer(model.EnvLayerSession).Set(joinKvName, string(data))
	return currCmdIdx, true
}

func JoinRun(
	argv model.ArgVals,
	cc *model.Cli,
	env *model.Env,
	flow *model.ParsedCmds,
	currCmdIdx int) (int, bool) {

	assertNotTailMode(flow, currCmdIdx)

	kvs := []joinKv{}
	kvsStr := env.GetRaw(joinKvName)
	if err := json.Unmarshal([]byte(kvsStr), &kvs); err != nil {
		panic(model.NewCmdError(flow.Cmds[currCmdIdx],
			fmt.Sprintf("parse argument '%s' failed: %s", kvsStr, err.Error())))
	}
	cmd := argv.GetRaw("cmd")
	if cmd == "" {
		panic(model.NewCmdError(flow.Cmds[currCmdIdx],
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
