package mod_meta

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

type ModMeta struct {
	sections SectionMap
}

func NewModMeta(path string) *ModMeta {
	meta := &ModMeta{
		sections: make(SectionMap),
	}
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("[NewModMeta] open mod meta file '%s' failed: %v", path, err))
	}
	meta.parse(contents, "\n", "=")
	return meta
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

	lines := bytes.Split(data, []byte(lineSep))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		size := len(line)
		if size == 0 {
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
