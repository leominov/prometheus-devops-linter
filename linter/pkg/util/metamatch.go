package util

import "regexp"

type MetaMatch struct {
	NameRaw  string `yaml:"name"`
	MatchRaw string `yaml:"match"`
	name     *regexp.Regexp
	match    *regexp.Regexp
}

func (m *MetaMatch) MatchTo(key, value string) bool {
	if !m.name.MatchString(key) {
		return true
	}
	if !m.match.MatchString(value) {
		return false
	}
	return true
}

func (m *MetaMatch) ProcessRegExp() error {
	re, err := regexp.Compile(m.MatchRaw)
	if err != nil {
		return err
	}
	m.match = re
	re, err = regexp.Compile(m.NameRaw)
	if err != nil {
		return err
	}
	m.name = re
	return nil
}
