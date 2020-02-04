package linter

import (
	"fmt"
	"path/filepath"

	"github.com/leominov/prometheus-devops-linter/linter/jobs"
	"github.com/leominov/prometheus-devops-linter/linter/rules"
)

type MetaLinter struct {
	c             *Config
	rulesLinter   *rules.Linter
	targetsLinter *jobs.Linter
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
	ml.targetsLinter = jobs.NewLinter(config.JobsConfig)
	return ml, nil
}

func (m *MetaLinter) LintFilesAs(linter string, paths []string) error {
	var filesToLint []string
	for _, path := range paths {
		files, err := filepath.Glob(path)
		if err != nil {
			return fmt.Errorf("Path error: %v", err)
		}
		filesToLint = append(filesToLint, files...)
	}
	switch linter {
	case rules.LinterName:
		return m.rulesLinter.LintFiles(filesToLint)
	case jobs.LinterName:
		return m.targetsLinter.LintFiles(filesToLint)
	default:
		return fmt.Errorf("Incorrect linter type: %s", linter)
	}
}
