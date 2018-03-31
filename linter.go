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
	MatchGroupName         string       `yaml:"matchGroupName"`
	MatchRuleAlert         string       `yaml:"matchRuleAlertName"`
	RequireGroupName       bool         `yaml:"requireGroupName"`
	RequireGroupRules      bool         `yaml:"requireGroupRules"`
	RequireRuleAlert       bool         `yaml:"requireRuleAlertName"`
	RequireRuleExpr        bool         `yaml:"requireRuleExpr"`
	RequireRuleLabels      []string     `yaml:"requireRuleLabels"`
	RequireRuleAnnotations []string     `yaml:"requireRuleAnnotations"`
	MatchRuleLabels        []*MetaMatch `yaml:"matchRuleLabels"`
	MatchRuleAnnotations   []*MetaMatch `yaml:"matchRuleAnnotations"`
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

type MetaMatch struct {
	Name        string `yaml:"name"`
	Match       string `yaml:"match"`
	matchRegExp *regexp.Regexp
}

func NewLinter(configFile string) (*Linter, error) {
	var (
		linter *Linter
		err    error
	)
	if len(configFile) > 0 {
		linter, err = NewLinterFromFile(configFile)
		if err != nil {
			return nil, err
		}
	} else {
		linter = &Linter{
			MatchGroupName:    "^([a-zA-Z0-9]+)$",
			MatchRuleAlert:    "^([a-zA-Z0-9]+)$",
			RequireGroupName:  true,
			RequireGroupRules: true,
			RequireRuleAlert:  true,
			RequireRuleExpr:   true,
			RequireRuleLabels: []string{
				"severity",
			},
			MatchRuleLabels:        []*MetaMatch{},
			RequireRuleAnnotations: []string{},
			MatchRuleAnnotations:   []*MetaMatch{},
		}
	}
	if err := linter.InitRegExpMatcher(); err != nil {
		return nil, err
	}
	return linter, nil
}

func NewLinterFromFile(configFile string) (*Linter, error) {
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	linter := &Linter{}
	err = yaml.Unmarshal(bytes, &linter)
	if err != nil {
		return nil, err
	}
	return linter, nil
}

func (l *Linter) InitRegExpMatcher() error {
	if len(l.MatchRuleAlert) > 0 {
		ruleAlertRegExp, err := regexp.Compile(l.MatchRuleAlert)
		if err != nil {
			return err
		}
		l.ruleAlertRegExp = ruleAlertRegExp
	}
	if len(l.MatchGroupName) > 0 {
		groupNameRegExp, err := regexp.Compile(l.MatchGroupName)
		if err != nil {
			return err
		}
		l.groupNameRegExp = groupNameRegExp
	}
	for _, labelMatch := range l.MatchRuleLabels {
		re, err := regexp.Compile(labelMatch.Match)
		if err != nil {
			return err
		}
		labelMatch.matchRegExp = re
	}
	for _, annMatch := range l.MatchRuleAnnotations {
		re, err := regexp.Compile(annMatch.Match)
		if err != nil {
			return err
		}
		annMatch.matchRegExp = re
	}
	return nil
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
			errs = append(errs, fmt.Errorf("Group name must match: %s", l.MatchGroupName))
		}
	}
	if l.RequireGroupRules && len(group.Rules) == 0 {
		errs = append(errs, fmt.Errorf("Rules for group '%s' is required", group.Name))
	}
	return errs
}

func (m *MetaMatch) MatchTo(key, value string) bool {
	if m.Name != key {
		return true
	}
	if !m.matchRegExp.MatchString(value) {
		return false
	}
	return true
}

func (l *Linter) MatchLabels(label, value string) []error {
	var errs []error
	for _, labelMatch := range l.MatchRuleLabels {
		if !labelMatch.MatchTo(label, value) {
			errs = append(errs, fmt.Errorf("Label value '%s' not match to: %s", label, labelMatch.Match))
		}
	}
	return errs
}

func (l *Linter) MatchAnnotations(annotation, value string) []error {
	var errs []error
	for _, annMatch := range l.MatchRuleAnnotations {
		if !annMatch.MatchTo(annotation, value) {
			errs = append(errs, fmt.Errorf("Annotation value '%s' not match to: %s", annotation, annMatch.Match))
		}
	}
	return errs
}

func (l *Linter) LintProjectRule(rule *Rule) []error {
	var errs []error
	if l.RequireRuleAlert && len(rule.Alert) == 0 {
		errs = append(errs, errors.New("Alert name is required"))
	}
	if l.ruleAlertRegExp != nil {
		if ok := l.ruleAlertRegExp.MatchString(rule.Alert); !ok {
			errs = append(errs, fmt.Errorf("Alert name must match: %s", l.MatchRuleAlert))
		}
	}
	if l.RequireRuleExpr && len(rule.Expr) == 0 {
		errs = append(errs, errors.New("Rule expr is required"))
	}
	if len(l.RequireRuleLabels) > 0 && len(rule.Labels) == 0 {
		errs = append(errs, errors.New("Rule labels is required"))
	}
	for _, requiredLabel := range l.RequireRuleLabels {
		val, ok := rule.Labels[requiredLabel]
		if !ok || len(val) == 0 {
			errs = append(errs, fmt.Errorf("Rule label '%s' is required and must be non-empty", requiredLabel))
		}
	}
	for label, value := range rule.Labels {
		matchErrors := l.MatchLabels(label, value)
		errs = append(errs, matchErrors...)
	}
	if len(l.RequireRuleAnnotations) > 0 && len(rule.Annotations) == 0 {
		errs = append(errs, errors.New("Rule annotations is required"))
	}
	for _, requiredAnnotation := range l.RequireRuleAnnotations {
		val, ok := rule.Annotations[requiredAnnotation]
		if !ok || len(val) == 0 {
			errs = append(errs, fmt.Errorf("Rule annotation '%s' is required and must be non-empty", requiredAnnotation))
		}
	}
	for annotation, value := range rule.Annotations {
		matchErrors := l.MatchAnnotations(annotation, value)
		errs = append(errs, matchErrors...)
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
