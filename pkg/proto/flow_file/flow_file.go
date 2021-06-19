package flow_file

import (
	"github.com/pingcap/ticat/pkg/proto/meta_file"
)

func LoadFlowFile(path string) (flow []string, help string, abbrs string) {
	meta := meta_file.NewMetaFile(path)
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
	if len(flow) != 0 {
		section.SetMultiLineVal("flow", flow)
	}
	meta.Save()
}
