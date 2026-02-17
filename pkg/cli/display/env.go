package display

import (
	"sort"
	"strings"

	"github.com/innerr/ticat/pkg/core/model"
)

func DumpEnvTree(screen model.Screen, env *model.Env, indentSize int) {
	lines, _ := dumpEnv(env, nil, true, true, true, true, nil, indentSize)
	for _, line := range lines {
		screen.Print(line + "\n")
	}
}

func DumpEssentialEnvFlattenVals(screen model.Screen, env *model.Env, findStrs ...string) {
	filterPrefixs := []string{
		"session",
		"strs.",
		"sys.",
		"display.",
	}
	flatten := env.Flatten(false, filterPrefixs, true)
	dumpEnvFlattenVals(screen, env, flatten, findStrs...)
}

func DumpEnvFlattenVals(screen model.Screen, env *model.Env, findStrs ...string) {
	flatten := env.Flatten(true, nil, true)
	dumpEnvFlattenVals(screen, env, flatten, findStrs...)
}

func KeyValueDisplayStr(key string, value string, env *model.Env) string {
	value = mayMaskSensitiveVal(env, key, value)
	return ColorKey(key, env) + ColorSymbol(" = ", env) + mayQuoteStr(value)
}

func KeyValueDisplayStrEx(key string, value string, env *model.Env, envKeysInfo *model.EnvKeysInfo) (string, int) {
	extraLen := ColorExtraLen(env, "symbol", "key")
	if envKeysInfo != nil {
		keyInfo := envKeysInfo.Get(key)
		if keyInfo != nil {
			if len(keyInfo.InvisibleDisplay) != 0 {
				value = keyInfo.InvisibleDisplay
			} else if keyInfo.DisplayLen != 0 {
				extraLen += len(value) - keyInfo.DisplayLen
			}
		}
	}
	value = mayMaskSensitiveVal(env, key, value)
	return ColorKey(key, env) + ColorSymbol(" = ", env) + mayQuoteStr(value), extraLen
}

func dumpEnvFlattenVals(screen model.Screen, env *model.Env, flatten map[string]string, findStrs ...string) {
	var keys []string
	for k := range flatten {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := flatten[k]
		if len(findStrs) != 0 {
			notMatched := false
			for _, findStr := range findStrs {
				if !strings.Contains(k, findStr) &&
					!strings.Contains(v, findStr) {
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
	env *model.Env,
	envKeysInfo *model.EnvKeysInfo,
	printEnvLayer bool,
	printDefEnv bool,
	printRuntimeEnv bool,
	printEnvStrs bool,
	filterPrefixs []string,
	indentSize int) (res []string, extraLens []int) {

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
		for k := range flatten {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := mayMaskSensitiveVal(env, k, flatten[k])
			res = append(res, ColorKey(k, env)+ColorSymbol(" = ", env)+v)
			extraLens = append(extraLens, ColorExtraLen(env, "key", "symbol"))
		}
	} else {
		dumpEnvLayer(env, env, envKeysInfo, printEnvLayer, printDefEnv, filterPrefixs, &res, &extraLens, indentSize, 0)
	}
	return
}

func dumpEnvLayer(
	env *model.Env,
	topEnv *model.Env,
	envKeysInfo *model.EnvKeysInfo,
	printEnvLayer bool,
	printDefEnv bool,
	filterPrefixs []string,
	res *[]string,
	extraLens *[]int,
	indentSize int,
	depth int) {

	if env.LayerType() == model.EnvLayerDefault && !printDefEnv {
		return
	}
	var output []string
	var outputExtraLens []int
	indent := rpt(" ", depth*indentSize)
	keys, _ := env.Pairs()
	sort.Strings(keys)
	for _, k := range keys {
		v := env.Get(k)
		filtered := false
		// Not filter default layer values
		for _, filterPrefix := range filterPrefixs {
			if len(filterPrefix) != 0 && strings.HasPrefix(k, filterPrefix) && env.LayerType() != model.EnvLayerDefault {
				filtered = true
				break
			}
		}
		if !filtered {
			kvStr, extraLen := KeyValueDisplayStrEx(k, v.Raw, topEnv, envKeysInfo)
			output = append(output, indent+"- "+kvStr)
			outputExtraLens = append(outputExtraLens, extraLen)
		}
	}
	if env.Parent() != nil {
		dumpEnvLayer(env.Parent(), topEnv, envKeysInfo, printEnvLayer, printDefEnv,
			filterPrefixs, &output, &outputExtraLens, indentSize, depth+1)
	}
	if len(output) != 0 {
		*res = append(*res, indent+ColorSymbol("["+env.LayerTypeName()+"]", topEnv))
		*extraLens = append(*extraLens, ColorExtraLen(topEnv, "symbol"))
		*res = append(*res, output...)
		*extraLens = append(*extraLens, outputExtraLens...)
	}
}
