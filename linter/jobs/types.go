package jobs

type Project struct {
	Jobs     []*Job
	Filename string
}

type Job struct {
	Name          string          `yaml:"job_name"`
	MetricsPath   string          `yaml:"metrics_path"`
	StaticConfigs []*StaticTarget `yaml:"static_configs"`
}

func (j *Job) String() string {
	if len(j.Name) > 0 {
		return j.Name
	}
	return "(unnamed job)"
}

type StaticTarget struct {
	Targets []string          `yaml:"targets"`
	Labels  map[string]string `yaml:"labels"`
}
