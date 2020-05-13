package rules

import "time"

type Project struct {
	Groups   []*Group `yaml:"groups"`
	Filename string   `yaml:"-"`
}

// ---
// apiVersion: monitoring.coreos.com/v1
// kind: PrometheusRule
// metadata:
//   name: kube-rules
// spec:
//   groups: []
type PrometheusOperatorProject struct {
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
