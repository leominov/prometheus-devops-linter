package linter

import (
	"fmt"
	"path/filepath"

	"github.com/leominov/prometheus-devops-linter/linter/rules"
	"github.com/leominov/prometheus-devops-linter/linter/targets"
)

type MetaLinter struct {
	c             *Config
	rulesLinter   *rules.Linter
	targetsLinter *targets.Linter
}

func NewMetaLinter(configFile string) (*MetaLinter, error) {
	config, err := NewConfig(configFile)
	if err != nil {
		return nil, err
	}
	ml := &MetaLinter{
		c: config,
	}
	ml.rulesLinter = rules.NewLinter(config.RulesConfig)
	ml.targetsLinter = targets.NewLinter(config.TargetsConfig)
	return ml, nil
}

func (m *MetaLinter) LintFilesAs(linter string, paths []string) error {
	var filesToLint []string
	for _, path := range paths {
		files, err := filepath.Glob(path)
		if err != nil {
			return fmt.Errorf("Path error: %v", err)
		}
		for _, filename := range files {
			filesToLint = append(filesToLint, filename)
		}
	}
	switch linter {
	case "rules":
		m.rulesLinter.LintFiles(filesToLint)
	case "targets":
		m.targetsLinter.LintFiles(filesToLint)
	}
	return nil
}
