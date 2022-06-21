package flow_file

import (
	"fmt"

	"github.com/pingcap/ticat/pkg/proto/meta_file"
)

func LoadFlowFile(path string) (flow []string, help string, abbrs string) {
	metas := meta_file.NewMetaFile(path)
	if len(metas) != 1 {
		panic(fmt.Errorf("can't load content for edit from a combined flow file"))
	}
	meta := metas[0].Meta

	section := meta.GetGlobalSection()
	help = section.Get("help")
	abbrs = section.Get("abbrs")
	flow = section.GetMultiLineVal("flow", false)
	return
}

func SaveFlowFile(path string, flow []string, help string, abbrs string) {
	meta := meta_file.CreateMetaFile(path)
	section := meta.GetGlobalSection()
	if len(help) != 0 {
		section.Set("help", help)
	}
	if len(abbrs) != 0 {
		section.Set("abbrs", abbrs)
	}
	section.Set("args.auto", "*")
	if len(flow) != 0 {
		section.SetMultiLineVal("flow", flow)
	}
	meta.Save()
}
