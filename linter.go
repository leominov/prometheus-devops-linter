package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type Linter struct {
	GroupNameRegExp        string
	RuleAlertRegExp        string
	RequireGroupName       bool
	RequireGroupRules      bool
	RequireRuleAlert       bool
	RequireRuleExpr        bool
	RequireRuleLabels      []string
	RequireRuleAnnotations []string
	groupNameRegExp        *regexp.Regexp
	ruleAlertRegExp        *regexp.Regexp
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

func NewLinter() (*Linter, error) {
	linter := &Linter{
		GroupNameRegExp:   "^([a-zA-Z]+)$",
		RuleAlertRegExp:   "^([a-zA-Z]+)$",
		RequireGroupName:  true,
		RequireGroupRules: true,
		RequireRuleAlert:  true,
		RequireRuleExpr:   true,
		RequireRuleLabels: []string{
			"env",
			"group",
			"severity",
		},
		RequireRuleAnnotations: []string{
			"description",
			"summary",
		},
	}
	if len(linter.RuleAlertRegExp) > 0 {
		ruleAlertRegExp, err := regexp.Compile(linter.RuleAlertRegExp)
		if err != nil {
			return nil, err
		}
		linter.ruleAlertRegExp = ruleAlertRegExp
	}
	if len(linter.GroupNameRegExp) > 0 {
		groupNameRegExp, err := regexp.Compile(linter.GroupNameRegExp)
		if err != nil {
			return nil, err
		}
		linter.groupNameRegExp = groupNameRegExp
	}
	return linter, nil
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
	if l.groupNameRegExp != nil {
		if ok := l.groupNameRegExp.MatchString(group.Name); !ok {
			errs = append(errs, fmt.Errorf("Group name must match: %s", l.GroupNameRegExp))
		}
	}
	if l.RequireGroupRules && len(group.Rules) == 0 {
		errs = append(errs, fmt.Errorf("Rules for group '%s' is required", group.Name))
	}
	return errs
}

func (l *Linter) LintProjectRule(rule *Rule) []error {
	var errs []error
	if l.RequireRuleAlert && len(rule.Alert) == 0 {
		errs = append(errs, errors.New("Alert name is requred"))
	}
	if l.ruleAlertRegExp != nil {
		if ok := l.ruleAlertRegExp.MatchString(rule.Alert); !ok {
			errs = append(errs, fmt.Errorf("Alert name must match: %s", l.RuleAlertRegExp))
		}
	}
	if l.RequireRuleExpr && len(rule.Expr) == 0 {
		errs = append(errs, errors.New("Alert expr is requred"))
	}
	if len(l.RequireRuleLabels) > 0 && len(rule.Labels) == 0 {
		errs = append(errs, errors.New("Alert labels is requred"))
	}
	for _, requiredLabel := range l.RequireRuleLabels {
		val, ok := rule.Labels[requiredLabel]
		if !ok || len(val) == 0 {
			errs = append(errs, fmt.Errorf("Alert label '%s' is requred and must be non-empty", requiredLabel))
		}
	}
	if len(l.RequireRuleAnnotations) > 0 && len(rule.Annotations) == 0 {
		errs = append(errs, errors.New("Alert annotations is requred"))
	}
	for _, requiredAnnotation := range l.RequireRuleAnnotations {
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
