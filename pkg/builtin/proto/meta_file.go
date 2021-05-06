package proto

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

func NewSection() Section {
	return Section {
		map[string]string{},
		[]string{},
	}
}

type Sections map[string]Section

type Meta struct {
	sections Sections
}

func (self *Meta) Get(key string) string {
	return self.SectionGet("", key)
}

func (self *Meta) SectionGet(sectionKey string, key string) (val string) {
	section, ok := self.sections[sectionKey]
	if ok {
		return section.Get(key)
	}
	return
}

func (self *Meta) GetSection(name string) (section Section, ok bool) {
	section, ok = self.sections[name]
	return section, ok
}

func NewMeta(path string, lineSep string, kvSep string) (*Meta, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	self := &Meta{Sections{}}
	section := NewSection()
	self.sections[""] = section

	for _, line := range bytes.Split(data, []byte(lineSep)) {
		line = bytes.TrimSpace(line)
		size := len(line)
		if size == 0 || line[0] == '#' {
			continue
		}
		if line[0] == '[' && line[size-1] == ']' {
			self.sections[string(line[1:size-1])] = NewSection()
			continue
		}
		pos := bytes.Index(line, []byte(kvSep))
		if pos < 0 {
			err := fmt.Errorf("%s: parse failed", string(line))
			return nil, err
		}
		k := bytes.TrimSpace(line[0:pos])
		v := bytes.Trim(bytes.TrimSpace(line[pos+len(kvSep):]), "'\"")
		section.Set(string(k), string(v))
	}
	return self, nil
}

type Section struct {
	pairs map[string]string
	orderedList []string
}

func (self *Section) Set(key string, val string) {
	_, ok := self.pairs[key]
	if !ok {
		self.orderedList = append(self.orderedList, key)
	}
	self.pairs[key] = val
}

func (self *Section) Get(key string) string {
	return self.pairs[key]
}

func (self *Section) Keys() []string {
	return self.orderedList
}
