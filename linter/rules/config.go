package rules

import (
	"regexp"

	"github.com/leominov/prometheus-devops-linter/linter/pkg/util"
)

type Config struct {
	OneGroupPerFile        bool              `yaml:"groupPerFile"`
	MatchGroupName         string            `yaml:"matchGroupName"`
	MatchRuleAlert         string            `yaml:"matchRuleAlertName"`
	RequireGroupName       bool              `yaml:"requireGroupName"`
	UniqueGroupName        bool              `yaml:"uniqueGroupName"`
	RequireGroupRules      bool              `yaml:"requireGroupRules"`
	RequireRuleAlert       bool              `yaml:"requireRuleAlertName"`
	RequireRuleExpr        bool              `yaml:"requireRuleExpr"`
	RequireRuleLabels      []string          `yaml:"requireRuleLabels"`
	RequireRuleAnnotations []string          `yaml:"requireRuleAnnotations"`
	MatchRuleLabels        []*util.MetaMatch `yaml:"matchRuleLabels"`
	MatchRuleAnnotations   []*util.MetaMatch `yaml:"matchRuleAnnotations"`
	groupNameRegExp        *regexp.Regexp
	ruleAlertRegExp        *regexp.Regexp
}

func (c *Config) SetDefaults() {
	c.MatchGroupName = "^([a-zA-Z0-9]+)$"
	c.MatchRuleAlert = "^([a-zA-Z0-9]+)$"
	c.RequireGroupName = true
	c.RequireGroupRules = true
	c.RequireRuleAlert = true
	c.RequireRuleExpr = true
	c.RequireRuleLabels = []string{
		"severity",
	}
	c.MatchRuleLabels = []*util.MetaMatch{}
	c.RequireRuleAnnotations = []string{}
	c.MatchRuleAnnotations = []*util.MetaMatch{}
}

func (c *Config) Init() error {
	if err := c.InitRegExpMatcher(); err != nil {
		return err
	}
	return nil
}

func (c *Config) InitRegExpMatcher() error {
	if len(c.MatchRuleAlert) > 0 {
		ruleAlertRegExp, err := regexp.Compile(c.MatchRuleAlert)
		if err != nil {
			return err
		}
		c.ruleAlertRegExp = ruleAlertRegExp
	}
	if len(c.MatchGroupName) > 0 {
		groupNameRegExp, err := regexp.Compile(c.MatchGroupName)
		if err != nil {
			return err
		}
		c.groupNameRegExp = groupNameRegExp
	}
	for _, labelMatch := range c.MatchRuleLabels {
		err := labelMatch.ProcessRegExp()
		if err != nil {
			return err
		}
	}
	for _, annMatch := range c.MatchRuleAnnotations {
		err := annMatch.ProcessRegExp()
		if err != nil {
			return err
		}
	}
	return nil
}
