package display

import (
	"sort"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DumpEnvTree(screen core.Screen, env *core.Env, indentSize int) {
	lines, _ := dumpEnv(env, true, true, true, true, nil, indentSize)
	for _, line := range lines {
		screen.Print(line + "\n")
	}
}

func DumpEssentialEnvFlattenVals(screen core.Screen, env *core.Env, findStrs ...string) {
	filterPrefixs := []string{
		"session",
		"strs.",
		"sys.",
		"display.",
	}
	flatten := env.Flatten(false, filterPrefixs, true)
	dumpEnvFlattenVals(screen, env, flatten, findStrs...)
}

func DumpEnvFlattenVals(screen core.Screen, env *core.Env, findStrs ...string) {
	flatten := env.Flatten(true, nil, true)
	dumpEnvFlattenVals(screen, env, flatten, findStrs...)
}

func KeyValueDisplayStr(key string, value string, env *core.Env) string {
	value = mayMaskSensitiveVal(key, value)
	return ColorKey(key, env) + ColorSymbol(" = ", env) + mayQuoteStr(value)
}

func dumpEnvFlattenVals(screen core.Screen, env *core.Env, flatten map[string]string, findStrs ...string) {
	var keys []string
	for k, _ := range flatten {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := flatten[k]
		if len(findStrs) != 0 {
			notMatched := false
			for _, findStr := range findStrs {
				if strings.Index(k, findStr) < 0 &&
					strings.Index(v, findStr) < 0 {
					notMatched = true
					break
				}
			}
			if notMatched {
				continue
			}
		}
		screen.Print(KeyValueDisplayStr(k, v, env) + "\n")
	}
}

func dumpEnv(
	env *core.Env,
	printEnvLayer bool,
	printDefEnv bool,
	printRuntimeEnv bool,
	printEnvStrs bool,
	filterPrefixs []string,
	indentSize int) (res []string, colored bool) {

	sep := env.Get("strs.env-path-sep").Raw
	if !printRuntimeEnv {
		sysPrefix := env.Get("strs.env-sys-path").Raw + sep
		filterPrefixs = append(filterPrefixs, sysPrefix)
	}
	if !printEnvStrs {
		strsPrefix := env.Get("strs.env-strs-path").Raw + sep
		filterPrefixs = append(filterPrefixs, strsPrefix)
	}

	if !printEnvLayer {
		flatten := env.Flatten(printDefEnv, filterPrefixs, true)
		var keys []string
		for k, _ := range flatten {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			res = append(res, ColorKey(k, env)+ColorSymbol(" = ", env)+flatten[k])
		}
		colored = true
	} else {
		dumpEnvLayer(env, printEnvLayer, printDefEnv, filterPrefixs, &res, indentSize, 0)
	}
	return
}

func dumpEnvLayer(
	env *core.Env,
	printEnvLayer bool,
	printDefEnv bool,
	filterPrefixs []string,
	res *[]string,
	indentSize int,
	depth int) {

	if env.LayerType() == core.EnvLayerDefault && !printDefEnv {
		return
	}
	var output []string
	indent := rpt(" ", depth*indentSize)
	keys, _ := env.Pairs()
	sort.Strings(keys)
	for _, k := range keys {
		v := env.Get(k)
		filtered := false
		for _, filterPrefix := range filterPrefixs {
			if len(filterPrefix) != 0 && strings.HasPrefix(k, filterPrefix) {
				filtered = true
				break
			}
		}
		if !filtered {
			raw := mayMaskSensitiveVal(k, v.Raw)
			output = append(output, indent+"- "+k+" = "+mayQuoteStr(raw))
		}
	}
	if env.Parent() != nil {
		dumpEnvLayer(env.Parent(), printEnvLayer, printDefEnv, filterPrefixs, &output, indentSize, depth+1)
	}
	if len(output) != 0 {
		*res = append(*res, indent+"["+env.LayerTypeName()+"]")
		*res = append(*res, output...)
	}
}
