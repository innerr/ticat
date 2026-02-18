package flow_file

import (
	"fmt"
	"strings"

	"github.com/innerr/ticat/pkg/mods/persist/meta_file"
)

func LoadFlowFile(path string) (flow []string, help string, abbrs string,
	trivial string, autoArgs string, packSub string, err error) {

	var metas []meta_file.VirtualMetaFile
	metas, err = meta_file.NewMetaFile(path)
	if err != nil {
		return
	}
	if len(metas) != 1 {
		err = fmt.Errorf("can't load content for edit from a combined flow file")
		return
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
	trivial string, autoArgs string, packSub string) error {

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
	return meta.Save()
}
