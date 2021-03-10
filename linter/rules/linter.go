package rules

import (
	"errors"
	"fmt"

	"github.com/leominov/prometheus-devops-linter/linter/pkg/util"
	"github.com/sirupsen/logrus"
)

const (
	LinterName = "rules"
)

type Linter struct {
	c             *Config
	groupNameList map[string]bool
}

func NewLinter(config *Config) *Linter {
	return &Linter{
		c:             config,
		groupNameList: make(map[string]bool),
	}
}

func (l *Linter) IsUniqueGroup(group *Group) bool {
	_, ok := l.groupNameList[group.Name]
	return !ok
}

func (l *Linter) LintProjectGroup(group *Group) []error {
	var errs []error
	if l.c.UniqueGroupName {
		if l.IsUniqueGroup(group) {
			l.groupNameList[group.Name] = true
		} else {
			errs = append(errs, errors.New("Group name must be unique"))
		}
	}
	if l.c.RequireGroupName && len(group.Name) == 0 {
		errs = append(errs, errors.New("Group name is required"))
	}
	if l.c.groupNameRegExp != nil {
		if ok := l.c.groupNameRegExp.MatchString(group.Name); !ok {
			errs = append(errs, fmt.Errorf("Group name must match: %s", l.c.MatchGroupName))
		}
	}
	if l.c.RequireGroupRules && len(group.Rules) == 0 {
		errs = append(errs, fmt.Errorf("Rules for group '%s' is required", group.Name))
	}
	return errs
}

func (l *Linter) MatchLabels(label, value string) []error {
	var errs []error
	for _, labelMatch := range l.c.MatchRuleLabels {
		if !labelMatch.MatchTo(label, value) {
			errs = append(errs, fmt.Errorf("Label value '%s' not match to: %s", label, labelMatch.MatchRaw))
		}
	}
	return errs
}

func (l *Linter) MatchAnnotations(annotation, value string) []error {
	var errs []error
	for _, annMatch := range l.c.MatchRuleAnnotations {
		if !annMatch.MatchTo(annotation, value) {
			errs = append(errs, fmt.Errorf("Annotation value '%s' not match to: %s", annotation, annMatch.MatchRaw))
		}
	}
	return errs
}

func (l *Linter) LintProjectRecord(record *Rule) []error {
	var errs []error
	if l.c.RequireRuleExpr && len(record.Expr) == 0 {
		errs = append(errs, errors.New("Record expr is required"))
	}
	return errs
}

func (l *Linter) LintProjectRule(rule *Rule) []error {
	var errs []error
	if l.c.RequireRuleAlert && len(rule.Alert) == 0 {
		errs = append(errs, errors.New("Alert name is required"))
	}
	if l.c.ruleAlertRegExp != nil {
		if ok := l.c.ruleAlertRegExp.MatchString(rule.Alert); !ok {
			errs = append(errs, fmt.Errorf("Alert name must match: %s", l.c.MatchRuleAlert))
		}
	}
	if l.c.RequireRuleExpr && len(rule.Expr) == 0 {
		errs = append(errs, errors.New("Rule expr is required"))
	}
	if len(l.c.RequireRuleLabels) > 0 && len(rule.Labels) == 0 {
		errs = append(errs, errors.New("Rule labels is required"))
	}
	for _, requiredLabel := range l.c.RequireRuleLabels {
		val, ok := rule.Labels[requiredLabel]
		if !ok || len(val) == 0 {
			errs = append(errs, fmt.Errorf("Rule label '%s' is required and must be non-empty", requiredLabel))
		}
	}
	for label, value := range rule.Labels {
		matchErrors := l.MatchLabels(label, value)
		errs = append(errs, matchErrors...)
	}
	if len(l.c.RequireRuleAnnotations) > 0 && len(rule.Annotations) == 0 {
		errs = append(errs, errors.New("Rule annotations is required"))
	}
	for _, requiredAnnotation := range l.c.RequireRuleAnnotations {
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

func (l *Linter) LintProject(project *Project) error {
	var withErrors bool
	if l.c.OneGroupPerFile && len(project.Groups) > 1 {
		withErrors = true
		util.PrintErrors(project.Filename, []error{errors.New("Allowed one group per file")})
	}
	for _, group := range project.Groups {
		groupErrors := l.LintProjectGroup(group)
		if len(groupErrors) > 0 {
			withErrors = true
			util.PrintErrors(group.String(), groupErrors)
		}
		for _, rule := range group.Rules {
			if len(rule.Record) != 0 {
				recordErrors := l.LintProjectRecord(rule)
				if len(recordErrors) > 0 {
					withErrors = true
					util.PrintErrors(fmt.Sprintf("%s > %s", group, rule), recordErrors)
				}
				continue
			}
			ruleErrors := l.LintProjectRule(rule)
			if len(ruleErrors) > 0 {
				withErrors = true
				util.PrintErrors(fmt.Sprintf("%s > %s", group, rule), ruleErrors)
			}
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
		logrus.Infof("Processing '%s' rules file...", filename)
		project, err := LoadProjectFromFile(filename)
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
