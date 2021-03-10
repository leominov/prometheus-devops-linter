package rules

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

type Project struct {
	Groups   []*Group `yaml:"groups"`
	Filename string   `yaml:"-"`

	// ---
	// apiVersion: monitoring.coreos.com/v1
	// kind: PrometheusRule
	// metadata:
	//   name: kube-rules
	// spec:
	//   groups: []
	Spec struct {
		Groups []*Group `yaml:"groups"`
	} `yaml:"spec"`
}

type Group struct {
	Name  string  `yaml:"name"`
	Rules []*Rule `yaml:"rules"`
}

func (g *Group) String() string {
	if len(g.Name) > 0 {
		return g.Name
	}
	return "(unnamed group)"
}

type Rule struct {
	Record      string
	Alert       string
	Expr        string
	For         time.Duration
	Labels      map[string]string
	Annotations map[string]string
}

func (r *Rule) String() string {
	if len(r.Alert) > 0 {
		return r.Alert
	}
	return "(unnamed rule)"
}

func LoadProjectFromFile(filename string) (*Project, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	project := &Project{
		Filename: filename,
	}
	err = yaml.Unmarshal(bytes, &project)
	if err != nil {
		return nil, err
	}
	if len(project.Spec.Groups) != 0 {
		project.Groups = project.Spec.Groups
		project.Spec.Groups = nil
		return project, err
	}
	return project, nil
}
