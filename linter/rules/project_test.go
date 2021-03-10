package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadProjectFromFile(t *testing.T) {
	_, err := LoadProjectFromFile("test_data/invalid.not-found.yaml")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "no such file")
	}

	_, err = LoadProjectFromFile("test_data/invalid.yaml")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "yaml")
	}

	p, err := LoadProjectFromFile("test_data/valid.yaml")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 0, len(p.Spec.Groups))
	assert.Equal(t, 1, len(p.Groups))
	assert.Equal(t, 3, len(p.Groups[0].Rules))

	p, err = LoadProjectFromFile("test_data/valid.prometheusrule.yaml")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 0, len(p.Spec.Groups))
	assert.Equal(t, 1, len(p.Groups))
	assert.Equal(t, 1, len(p.Groups[0].Rules))
}
