package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type Linter struct {
	ShowAllErrors           bool
	RequireGroupName        bool
	RequireGroupRules       bool
	RequireAlertAlert       bool
	RequireAlertExpr        bool
	RequireAlertLabels      []string
	RequireAlertAnnotations []string
}

type Project struct {
	Groups []*Group `yaml:"groups"`
}

type Group struct {
	Name  string  `yaml:"name"`
	Rules []*Rule `yaml:"rules"`
}

type Rule struct {
	Alert       string
	Expr        string
	For         time.Duration
	Labels      map[string]string
	Annotations map[string]string
}

func NewLinter() *Linter {
	return &Linter{
		ShowAllErrors:     true,
		RequireGroupName:  true,
		RequireGroupRules: true,
		RequireAlertAlert: true,
		RequireAlertExpr:  true,
		RequireAlertLabels: []string{
			"env",
			"group",
			"severity",
		},
		RequireAlertAnnotations: []string{
			"description",
			"summary",
		},
	}
}

func (g *Group) String() string {
	if len(g.Name) > 0 {
		return g.Name
	}
	return "(unnamed group)"
}

func (r *Rule) String() string {
	if len(r.Alert) > 0 {
		return r.Alert
	}
	return "(unnamed alert)"
}

func (l *Linter) LintProjectGroup(group *Group) []error {
	var errs []error
	if l.RequireGroupName && len(group.Name) == 0 {
		errs = append(errs, errors.New("Group name is required"))
	}
	if l.RequireGroupRules && len(group.Rules) == 0 {
		errs = append(errs, fmt.Errorf("Rules for group '%s' is required", group.Name))
	}
	return errs
}

func (l *Linter) LintProjectRule(rule *Rule) []error {
	var errs []error
	if l.RequireAlertAlert && len(rule.Alert) == 0 {
		errs = append(errs, errors.New("Alert name is requred"))
	}
	if l.RequireAlertExpr && len(rule.Expr) == 0 {
		errs = append(errs, errors.New("Alert expr is requred"))
	}
	if len(l.RequireAlertLabels) > 0 && len(rule.Labels) == 0 {
		errs = append(errs, errors.New("Alert labels is requred"))
	}
	for _, requiredLabel := range l.RequireAlertLabels {
		val, ok := rule.Labels[requiredLabel]
		if !ok || len(val) == 0 {
			errs = append(errs, fmt.Errorf("Alert label '%s' is requred and must be non-empty", requiredLabel))
		}
	}
	if len(l.RequireAlertAnnotations) > 0 && len(rule.Annotations) == 0 {
		errs = append(errs, errors.New("Alert annotations is requred"))
	}
	for _, requiredAnnotation := range l.RequireAlertAnnotations {
		val, ok := rule.Annotations[requiredAnnotation]
		if !ok || len(val) == 0 {
			errs = append(errs, fmt.Errorf("Alert annotation '%s' is requred and must be non-empty", requiredAnnotation))
		}
	}
	return errs
}

func PrintErrors(pref string, errs []error) {
	for _, err := range errs {
		messagePref := pref
		if len(messagePref) > 0 {
			messagePref = fmt.Sprintf("%s: ", messagePref)
		}
		logrus.Errorf("%s%s", messagePref, err.Error())
	}
}

func (l *Linter) LintProject(project *Project) error {
	var withErrors bool
	for _, group := range project.Groups {
		groupErrors := l.LintProjectGroup(group)
		if len(groupErrors) > 0 {
			withErrors = true
			PrintErrors(group.String(), groupErrors)
		}
		for _, rule := range group.Rules {
			ruleErrors := l.LintProjectRule(rule)
			if len(ruleErrors) > 0 {
				withErrors = true
				PrintErrors(fmt.Sprintf("%s > %s", group, rule), ruleErrors)
			}
		}
	}
	if withErrors {
		return errors.New("Project with errors")
	}
	return nil
}

func (l *Linter) LoadProjectFromFile(filename string) (*Project, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	project := &Project{}
	err = yaml.Unmarshal(bytes, &project)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (l *Linter) ProcessFilesFromPath(p string) error {
	var doneWithErrors bool
	files, err := filepath.Glob(p)
	if err != nil {
		return fmt.Errorf("Path error: %v", err)
	}
	for _, filename := range files {
		logrus.Infof("Processing '%s' file...", filename)
		project, err := l.LoadProjectFromFile(filename)
		if err != nil {
			return err
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
