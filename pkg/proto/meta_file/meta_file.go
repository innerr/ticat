package meta_file

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type MetaFile struct {
	sections    SectionMap
	orderedKeys []string
	path        string
	lineSep     string
	kvSep       string
}

func NewMetaFile(path string) *MetaFile {
	meta, err := NewMetaFileEx(path)
	if err != nil {
		panic(fmt.Errorf("[NewMetaFile] open mod meta file '%s' failed: %v", path, err))
	}
	return meta
}

func NewMetaFileEx(path string) (meta *MetaFile, err error) {
	meta = CreateMetaFile(path)
	var contents []byte
	contents, err = ioutil.ReadFile(path)
	if err == nil {
		meta.parse(contents)
	}
	return
}

func CreateMetaFile(path string) (meta *MetaFile) {
	meta = &MetaFile{
		make(SectionMap),
		nil,
		path,
		LineSep,
		KvSep,
	}
	return
}

func (self *MetaFile) Path() string {
	return self.path
}

func (self *MetaFile) Get(key string) string {
	return self.SectionGet(GlobalSectionName, key)
}

func (self *MetaFile) SectionGet(sectionName string, key string) (val string) {
	section := self.sections[sectionName]
	if section != nil {
		return section.Get(key)
	}
	return
}

func (self *MetaFile) GetSection(key string) *Section {
	section, _ := self.sections[key]
	return section
}

func (self *MetaFile) GetGlobalSection() *Section {
	return self.NewOrGetSection(GlobalSectionName)
}

func (self *MetaFile) NewOrGetSection(key string) *Section {
	section, ok := self.sections[key]
	if ok {
		return section
	}
	section = NewSection()
	self.sections[key] = section
	self.orderedKeys = append(self.orderedKeys, key)
	return section
}

func (self *MetaFile) GetAll() SectionMap {
	return self.sections
}

func (self *MetaFile) parse(data []byte) {
	var sectionName string
	section := NewSection()
	self.sections[sectionName] = section

	var multiLineKey string
	var multiLineValue []string

	tryAppendMultiLine := func(line string) bool {
		if len(multiLineKey) == 0 {
			return false
		}
		multiLineFinish := false
		if line[len(line)-1] == MultiLineBreaker {
			line = line[:len(line)-1]
		} else {
			multiLineFinish = true
		}
		multiLineValue = append(multiLineValue, strings.TrimSpace(line))
		if multiLineFinish {
			section.Set(multiLineKey, strings.Join(multiLineValue, "\n"))
			multiLineKey = ""
			multiLineValue = nil
		}
		return true
	}

	checkMultiLineStart := func(k string, v string) bool {
		if len(v) == 0 {
			return false
		}
		if v[len(v)-1] != MultiLineBreaker {
			return false
		}
		v = strings.TrimSpace(v[:len(v)-1])
		multiLineKey = k
		if len(v) != 0 {
			multiLineValue = append(multiLineValue, v)
		}
		return true
	}

	// TODO: convert to string too many times
	lines := bytes.Split(data, []byte(self.lineSep))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		size := len(line)
		if size == 0 {
			continue
		}
		if tryAppendMultiLine(string(line)) {
			continue
		}
		if line[0] == CommentPrefix {
			continue
		}
		if line[0] == SectionBracketLeft && line[size-1] == SectionBracketRight {
			sectionName = string(line[1 : size-1])
			section = NewSection()
			self.sections[sectionName] = section
			continue
		}

		pos := bytes.Index(line, []byte(self.kvSep))
		if pos < 0 {
			panic(fmt.Errorf("[MetaFile.parse] bad kv format: %s", line))
		}

		k := bytes.TrimSpace(line[0:pos])
		v := bytes.TrimSpace(line[pos+len(self.kvSep):])
		if checkMultiLineStart(string(k), string(v)) {
			continue
		}
		section.Set(string(k), string(v))
	}
}

func (self *MetaFile) Save() {
	file, err := os.OpenFile(self.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(fmt.Errorf("[MetaFile.Save] open mod meta file '%s' for save failed: %v",
			self.path, err))
	}
	defer file.Close()
	self.save(file)
}

func (self *MetaFile) save(w io.Writer) {
	saveKey := func(key string, val string) (multiLine bool) {
		lines := strings.Split(val, self.lineSep)
		if len(lines) == 1 {
			fmt.Fprintf(w, "%s %s %s\n", key, self.kvSep, val)
			return false
		} else {
			fmt.Fprintf(w, "%s %s %s %c\n", key, self.kvSep, lines[0], MultiLineBreaker)
			lines = lines[1:]
			for i, line := range lines {
				if i != len(lines)-1 {
					fmt.Fprintf(w, "    %s %c\n", line, MultiLineBreaker)
				} else {
					fmt.Fprintf(w, "    %s\n", line)
				}
			}
			return true
		}
	}

	for i, name := range self.orderedKeys {
		section := self.sections[name]
		keys := section.Keys()
		if len(name) != 0 {
			fmt.Fprintf(w, "%c%s%c\n", SectionBracketLeft, name, SectionBracketRight)
		}
		for j, key := range keys {
			multiLine := saveKey(key, section.Get(key))
			if multiLine && j != len(keys)-1 {
				fmt.Fprintf(w, "\n")
			}
		}
		if i != len(self.orderedKeys)-1 {
			fmt.Fprintf(w, "\n")
		}
	}
}

type SectionMap map[string]*Section

type Section struct {
	pairs       map[string]string
	orderedKeys []string
}

func NewSection() *Section {
	return &Section{
		map[string]string{},
		[]string{},
	}
}

func (self *Section) Get(key string) string {
	val, _ := self.pairs[key]
	return strings.Trim(val, ValTrimChars)
}

func (self *Section) GetUnTrim(key string) string {
	val, _ := self.pairs[key]
	return val
}

func (self *Section) GetMultiLineVal(key string, trim bool) []string {
	val, ok := self.pairs[key]
	if !ok {
		return nil
	}
	if trim {
		val = strings.Trim(val, ValTrimChars)
	}
	return strings.Split(val, LineSep)
}

func (self *Section) Keys() []string {
	return self.orderedKeys
}

func (self *Section) Set(key string, val string) {
	_, ok := self.pairs[key]
	self.pairs[key] = val
	if !ok {
		self.orderedKeys = append(self.orderedKeys, key)
	}
	return
}

func (self *Section) SetMultiLineVal(key string, val []string) {
	self.Set(key, strings.Join(val, LineSep))
}

const (
	LineSep           = "\n"
	KvSep             = "="
	GlobalSectionName = ""
	ValTrimChars      = "'\""

	MultiLineBreaker    = '\\'
	CommentPrefix       = '#'
	SectionBracketLeft  = '['
	SectionBracketRight = ']'
)
