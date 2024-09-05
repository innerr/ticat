package display

import (
	"sort"

	"github.com/innerr/ticat/pkg/core/model"
)

// TODO: list cmd usage, now only list the tags
func ListTags(
	cmds *model.CmdTree,
	screen model.Screen,
	env *model.Env) {

	tags := NewTags()
	listTags(cmds, screen, env, tags)

	names := tags.Names()
	sort.Strings(names)
	tagMark := env.GetRaw("strs.tag-mark")
	for _, name := range names {
		screen.Print(ColorTag(tagMark+name, env) + "\n")
	}
}

type Tags struct {
	names []string
	set   map[string]bool
}

func NewTags() *Tags {
	return &Tags{nil, map[string]bool{}}
}

func (self *Tags) Add(name string) {
	if self.set[name] {
		return
	}
	self.set[name] = true
	self.names = append(self.names, name)
}

func (self *Tags) Names() []string {
	return self.names
}

func listTags(
	cmd *model.CmdTree,
	screen model.Screen,
	env *model.Env,
	tags *Tags) {

	for _, tag := range cmd.Tags() {
		tags.Add(tag)
	}
	for _, name := range cmd.SubNames() {
		listTags(cmd.GetSub(name), screen, env, tags)
	}
}
