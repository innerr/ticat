package flow_file

import (
	"fmt"
	"strings"

	"github.com/innerr/ticat/pkg/mods/persist/meta_file"
)

func LoadFlowFile(path string) (flow []string, help string, abbrs string,
	trivial string, autoArgs string, packSub string) {

	metas := meta_file.NewMetaFile(path)
	if len(metas) != 1 {
		panic(fmt.Errorf("can't load content for edit from a combined flow file"))
	}
	meta := metas[0].Meta

	section := meta.GetGlobalSection()
	help = section.Get("help")
	trivial = section.Get("trivial")
	abbrs = section.Get("abbrs")
	if len(abbrs) == 0 {
		abbrs = section.Get("abbr")
	}
	autoArgs = section.Get("args.auto")
	packSub = section.Get("pack-subflow")
	if len(packSub) == 0 {
		packSub = section.Get("pack-sub")
	}
	flow = section.GetMultiLineVal("flow", false)
	return
}

func SaveFlowFile(path string, flow []string, help string, abbrs string,
	trivial string, autoArgs string, packSub string) {

	meta := meta_file.CreateMetaFile(path)
	section := meta.GetGlobalSection()
	if len(help) != 0 {
		section.Set("help", help)
	}
	trivial = strings.TrimSpace(trivial)
	if len(trivial) != 0 && trivial != "0" {
		section.Set("trivial", trivial)
	}
	if len(abbrs) != 0 {
		section.Set("abbrs", abbrs)
	}
	if len(autoArgs) == 0 {
		autoArgs = "*"
	}
	if len(packSub) != 0 {
		section.Set("pack-subflow", packSub)
	}
	section.Set("args.auto", autoArgs)
	if len(flow) != 0 {
		section.SetMultiLineVal("flow", flow)
	}
	meta.Save()
}
