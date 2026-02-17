package meta_file

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type VirtualMetaFile struct {
	Meta        *MetaFile
	VirtualPath string
	NotVirtual  bool
}

type MetaFile struct {
	sections    SectionMap
	orderedKeys []string
	path        string
	lineSep     string
	kvSep       string
}

func NewMetaFile(path string) []VirtualMetaFile {
	metas, err := NewMetaFileEx(path)
	if err != nil {
		panic(fmt.Errorf("[NewMetaFile] open mod meta file '%s' failed: %v", path, err))
	}
	return metas
}

func NewMetaFileEx(path string) (metas []VirtualMetaFile, err error) {
	var content []byte
	content, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}
	paths, contents, notVirtuals := parseCombinedFile(path, content, LineSep)
	for i, it := range contents {
		meta := CreateMetaFile(paths[i])
		meta.parse(it)
		metas = append(metas, VirtualMetaFile{meta, paths[i], notVirtuals[i]})
	}
	return
}

func parseCombinedFile(path string, data []byte, lineSep string) (paths []string, contents [][]string, notVirtuals []bool) {
	notVirtual := true
	currPath := path
	currLines := []string{}
	raw := bytes.Split(data, []byte(lineSep))
	for _, lineBytes := range raw {
		line := string(bytes.TrimSpace(lineBytes))
		if strings.HasPrefix(line, CombinedFileHint) {
			cand := strings.TrimSpace(line[len(CombinedFileHint):])
			if strings.HasPrefix(cand, CombinedFilePrefix1) {
				cand = strings.TrimSpace(cand[len(CombinedFilePrefix1):])
				if strings.HasPrefix(cand, CombinedFilePrefix2) {
					if !(currPath == path && len(currLines) == 0) {
						paths = append(paths, currPath)
						contents = append(contents, currLines)
						notVirtuals = append(notVirtuals, notVirtual)
					}
					currPath = strings.TrimSpace(cand[len(CombinedFilePrefix2):])
					currLines = []string{}
					notVirtual = false
				}
			}
		}
		currLines = append(currLines, line)
	}
	paths = append(paths, currPath)
	contents = append(contents, currLines)
	notVirtuals = append(notVirtuals, notVirtual)
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

func (self *MetaFile) KeysWithPrefix(keyPrefix string) (keys []string) {
	section := self.sections[GlobalSectionName]
	if section != nil {
		return section.KeysWithPrefix(keyPrefix)
	}
	return
}

func (self *MetaFile) SectionGet(sectionName string, key string) (val string) {
	section := self.sections[sectionName]
	if section != nil {
		return section.Get(key)
	}
	return
}

func (self *MetaFile) GetSection(key string) *Section {
	section := self.sections[key]
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

func (self *MetaFile) parse(lines []string) {
	var sectionName string
	section := NewSection()
	global := section
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

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		line = strings.TrimSpace(line)
		size := len(line)
		if size == 0 {
			continue
		}
		if tryAppendMultiLine(line) {
			continue
		}
		if line[0] == CommentPrefix {
			continue
		}

		if line[0] == SectionBracketLeft && line[size-1] == SectionBracketRight &&
			// Ignore [[...]]
			!(len(line) >= 4 && line[1] == SectionBracketLeft && line[size-2] == SectionBracketRight) {

			if len(line) > 2 && line[size-2] == '/' {
				k := line[1 : size-2]
				v := []string{}
				for i += 1; i < len(lines); i++ {
					line := lines[i]
					line = strings.TrimSpace(line)
					// Keep the comments and blank lines in this format
					//if line[0] == CommentPrefix {
					//	continue
					//}
					if len(line) == 0 {
						if len(v) != 0 {
							v = append(v, "")
						}
						continue
					}
					if line[0] == SectionBracketLeft && line[len(line)-1] == SectionBracketRight &&
						len(line) > 2 && line[1] == '/' {
						break
					}
					l := strings.TrimSpace(line)
					v = append(v, l)
				}
				if len(v) > 0 && len(v[len(v)-1]) == 0 {
					v = v[:len(v)-1]
				}
				global.SetMultiLineVal(k, v)
			} else {
				sectionName = line[1 : size-1]
				section = NewSection()
				self.sections[sectionName] = section
			}
			continue
		}

		pos := strings.Index(line, self.kvSep)
		if pos < 0 {
			panic(fmt.Errorf("[MetaFile.parse] bad kv format: %s", line))
		}

		k := strings.TrimSpace(line[0:pos])
		v := strings.TrimSpace(line[pos+len(self.kvSep):])
		if checkMultiLineStart(k, v) {
			continue
		}
		section.Set(k, v)
	}
}

func (self *MetaFile) Save() {
	file, err := os.OpenFile(self.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(fmt.Errorf("[MetaFile.Save] open mod meta file '%s' for save failed: %v",
			self.path, err))
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(fmt.Errorf("[MetaFile.Save] close mod meta file '%s' failed: %v",
				self.path, err))
		}
	}()
	self.save(file)
}

func (self *MetaFile) save(w io.Writer) {
	saveKey := func(key string, val string) (multiLine bool) {
		lines := strings.Split(val, self.lineSep)
		if len(lines) == 1 {
			if _, err := fmt.Fprintf(w, "%s %s %s\n", key, self.kvSep, val); err != nil {
				panic(fmt.Errorf("[MetaFile.save] write failed: %v", err))
			}
			return false
		} else {
			if _, err := fmt.Fprintf(w, "%s %s %s %c\n", key, self.kvSep, lines[0], MultiLineBreaker); err != nil {
				panic(fmt.Errorf("[MetaFile.save] write failed: %v", err))
			}
			lines = lines[1:]
			for i, line := range lines {
				if i != len(lines)-1 {
					if _, err := fmt.Fprintf(w, "    %s %c\n", line, MultiLineBreaker); err != nil {
						panic(fmt.Errorf("[MetaFile.save] write failed: %v", err))
					}
				} else {
					if _, err := fmt.Fprintf(w, "    %s\n", line); err != nil {
						panic(fmt.Errorf("[MetaFile.save] write failed: %v", err))
					}
				}
			}
			return true
		}
	}

	for i, name := range self.orderedKeys {
		section := self.sections[name]
		keys := section.Keys()
		if len(name) != 0 {
			if _, err := fmt.Fprintf(w, "%c%s%c\n", SectionBracketLeft, name, SectionBracketRight); err != nil {
				panic(fmt.Errorf("[MetaFile.save] write failed: %v", err))
			}
		}
		for j, key := range keys {
			multiLine := saveKey(key, section.GetUnTrim(key))
			if multiLine && j != len(keys)-1 {
				if _, err := fmt.Fprintf(w, "\n"); err != nil {
					panic(fmt.Errorf("[MetaFile.save] write failed: %v", err))
				}
			}
		}
		if i != len(self.orderedKeys)-1 {
			if _, err := fmt.Fprintf(w, "\n"); err != nil {
				panic(fmt.Errorf("[MetaFile.save] write failed: %v", err))
			}
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
	val := self.pairs[key]
	return strings.Trim(val, ValTrimChars)
}

func (self *Section) KeysWithPrefix(keyPrefix string) (keys []string) {
	for _, k := range self.orderedKeys {
		if strings.HasPrefix(k, keyPrefix) {
			keys = append(keys, k)
		}
	}
	return
}

func (self *Section) GetUnTrim(key string) string {
	val := self.pairs[key]
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
	res := strings.Split(val, LineSep)
	for len(res) > 0 && len(res[len(res)-1]) == 0 {
		res = res[:len(res)-1]
	}
	return res
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

	CombinedFileHint    = "###"
	CombinedFilePrefix1 = "file"
	CombinedFilePrefix2 = ":"
)
