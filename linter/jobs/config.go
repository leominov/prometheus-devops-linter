package jobs

type Config struct {
	UniqueJobName       bool     `yaml:"uniqueJobName"`
	UniqueTarget        bool     `yaml:"uniqueTarget"`
	RequireTargetLabels []string `yaml:"requireTargetLabels"`
}

func (c *Config) SetDefaults() {
	c.UniqueJobName = true
	c.UniqueTarget = true
}
