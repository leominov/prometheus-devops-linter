package linter

import (
	"io/ioutil"

	"github.com/leominov/prometheus-devops-linter/linter/rules"
	"github.com/leominov/prometheus-devops-linter/linter/targets"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	RulesConfig   *rules.Config   `yaml:"rules"`
	TargetsConfig *targets.Config `yaml:"targets"`
}

func loadConfigFromFile(configFile string) (*Config, error) {
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func NewConfig(path string) (*Config, error) {
	var (
		config *Config
		err    error
	)
	if len(path) > 0 {
		config, err = loadConfigFromFile(path)
		if err != nil {
			return nil, err
		}
	}
	config.SetDefaults()
	if err := config.RulesConfig.Init(); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *Config) SetDefaults() {
	if c.RulesConfig == nil {
		c.RulesConfig = &rules.Config{}
		c.RulesConfig.SetDefaults()
	}
	if c.TargetsConfig == nil {
		c.TargetsConfig = &targets.Config{}
		c.TargetsConfig.SetDefaults()
	}
}