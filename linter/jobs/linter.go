package jobs

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/leominov/prometheus-devops-linter/linter/pkg/util"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

const (
	LinterName = "jobs"
)

type Linter struct {
	c             *Config
	jobNameList   map[string]bool
	jobTargetList map[string]bool
}

func NewLinter(config *Config) *Linter {
	return &Linter{
		c:             config,
		jobNameList:   make(map[string]bool),
		jobTargetList: make(map[string]bool),
	}
}

func (l *Linter) LoadProjectFromFile(filename string) (*Project, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	project := &Project{
		Jobs:     []*Job{},
		Filename: filename,
	}
	err = yaml.Unmarshal(bytes, &project.Jobs)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (l *Linter) IsUniqueJob(job *Job) bool {
	_, ok := l.jobNameList[job.Name]
	if ok {
		return false
	}
	return true
}

func (l *Linter) IsUniqueJobTarget(target string) bool {
	_, ok := l.jobTargetList[target]
	if ok {
		return false
	}
	return true
}

func (l *Linter) LintJob(job *Job) []error {
	var errs []error
	if l.c.UniqueJobName {
		if l.IsUniqueJob(job) {
			l.jobNameList[job.Name] = true
		} else {
			errs = append(errs, errors.New("Job name must be unique"))
		}
	}
	if l.c.UniqueTarget {
		for _, targers := range job.StaticConfigs {
			for _, targer := range targers.Targets {
				fullTargetName := fmt.Sprintf("%s:%s", job.MetricsPath, targer)
				if l.IsUniqueJobTarget(fullTargetName) {
					l.jobTargetList[fullTargetName] = true
				} else {
					errs = append(errs, fmt.Errorf("Job target must be unique, found duplicate of %s", targer))
				}
			}
		}
	}
	if len(l.c.RequireTargetLabels) > 0 {
		for _, targets := range job.StaticConfigs {
			for _, requiredLabel := range l.c.RequireTargetLabels {
				val, ok := targets.Labels[requiredLabel]
				if !ok || len(val) == 0 {
					errs = append(errs, fmt.Errorf("Target label '%s' is required and must be non-empty", requiredLabel))
				}
			}
		}
	}
	return errs
}

func (l *Linter) LintProject(project *Project) error {
	var withErrors bool
	for _, job := range project.Jobs {
		jobErrors := l.LintJob(job)
		if len(jobErrors) > 0 {
			withErrors = true
			util.PrintErrors(job.String(), jobErrors)
		}
	}
	if withErrors {
		return errors.New("Project with errors")
	}
	return nil
}

func (l *Linter) LintFiles(files []string) error {
	var doneWithErrors bool
	for _, filename := range files {
		logrus.Infof("Processing '%s' jobs file...", filename)
		project, err := l.LoadProjectFromFile(filename)
		if err != nil {
			return err
		}
		if project == nil {
			logrus.Warnf("File %s is empty", filename)
			continue
		}
		if err := l.LintProject(project); err != nil {
			doneWithErrors = true
		}
	}
	if doneWithErrors {
		return errors.New("Done with errors")
	}
	return nil
}
