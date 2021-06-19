package meta_file

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
)

type ModMeta struct {
	sections SectionMap
}

func NewModMeta(path string) *ModMeta {
	meta, err := NewModMetaEx(path)
	if err != nil {
		panic(fmt.Errorf("[NewModMeta] open mod meta file '%s' failed: %v", path, err))
	}
	return meta
}

func NewModMetaEx(path string) (meta *ModMeta, err error) {
	meta = &ModMeta{
		sections: make(SectionMap),
	}
	var contents []byte
	contents, err = ioutil.ReadFile(path)
	if err == nil {
		meta.parse(contents, "\n", "=")
	}
	return
}

func (self *ModMeta) Get(key string) string {
	return self.SectionGet("", key)
}

func (self *ModMeta) SectionGet(sectionName string, key string) (val string) {
	session := self.sections[sectionName]
	if session != nil {
		return session.Get(key)
	}
	return
}

func (self *ModMeta) GetSession(key string) *Session {
	session, _ := self.sections[key]
	return session
}

func (self *ModMeta) GetAll() SectionMap {
	return self.sections
}

func (self *ModMeta) parse(data []byte, lineSep, kvSep string) {
	var sectionName string
	session := NewSession()
	self.sections[sectionName] = session

	var multiLineKey string
	var multiLineValue []string

	tryAppendMultiLine := func(line string) bool {
		if len(multiLineKey) == 0 {
			return false
		}
		multiLineFinish := false
		if line[len(line)-1] == '\\' {
			line = line[:len(line)-1]
		} else {
			multiLineFinish = true
		}
		multiLineValue = append(multiLineValue, strings.TrimSpace(line))
		if multiLineFinish {
			session.Set(multiLineKey, strings.Join(multiLineValue, "\n"))
			multiLineKey = ""
			multiLineValue = nil
		}
		return true
	}

	checkMultiLineStart := func(k string, v string) bool {
		if len(v) == 0 {
			return false
		}
		if v[len(v)-1] != '\\' {
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
	lines := bytes.Split(data, []byte(lineSep))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		size := len(line)
		if size == 0 {
			continue
		}
		if tryAppendMultiLine(string(line)) {
			continue
		}
		if line[0] == ';' || line[0] == '#' {
			continue
		}
		if line[0] == '[' && line[size-1] == ']' {
			sectionName = string(line[1 : size-1])
			session = NewSession()
			self.sections[sectionName] = session
			continue
		}

		pos := bytes.Index(line, []byte(kvSep))
		if pos < 0 {
			panic(fmt.Errorf("[ModMeta.parse] bad kv format: %s", line))
		}

		k := bytes.TrimSpace(line[0:pos])
		v := bytes.TrimSpace(line[pos+len(kvSep):])
		if checkMultiLineStart(string(k), string(v)) {
			continue
		}
		v = bytes.Trim(v, "'\"")
		session.Set(string(k), string(v))
	}
}

type SectionMap map[string]*Session

type Session struct {
	pairs       map[string]string
	orderedKeys []string
}

func NewSession() *Session {
	return &Session{
		map[string]string{},
		[]string{},
	}
}

func (self *Session) Get(key string) string {
	val, _ := self.pairs[key]
	return val
}

func (self *Session) Keys() []string {
	return self.orderedKeys
}

func (self *Session) Set(key string, val string) {
	_, ok := self.pairs[key]
	self.pairs[key] = val
	if !ok {
		self.orderedKeys = append(self.orderedKeys, key)
	}
	return
}
