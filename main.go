package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/leominov/prometheus-devops-linter/linter"
	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
)

const (
	DefaultConfigFilename = ".prometheus-linter.yaml"
)

var (
	AlertsPath = flag.String("path", "", "Directory with alert files")
	ConfigFile = flag.String("config-file", "", "Configuration file")
)

func DiscoverConfigFile() string {
	configEnv := os.Getenv("PROM_LINTER_CONFIG")
	if len(configEnv) > 0 {
		return configEnv
	}
	dir, err := os.Getwd()
	if err != nil {
		dir = "./"
	}
	filename := path.Join(dir, DefaultConfigFilename)
	if _, err := os.Stat(filename); err == nil {
		return filename
	}
	return ""
}

func ParseArgs() (string, []string, error) {
	args := os.Args
	if len(args) < 1 {
		return "", []string{}, errors.New("Linter type must be specified")
	}
	if len(args) < 2 {
		return "", []string{}, errors.New("Directory must be specified")
	}
	linterType := strings.ToLower(args[1])
	paths := args[2:]
	if linterType != "rules" && linterType != "targets" {
		return "", []string{}, fmt.Errorf("Incorrect linter type: %s", linterType)
	}
	return linterType, paths, nil
}

func main() {
	var configFile string
	configFile = *ConfigFile
	if len(configFile) == 0 {
		configFile = DiscoverConfigFile()
	}
	logrus.Infof("Starting prometheus-devops-linter %s...", version.Info())
	logrus.Infof("Configuration path: %s", configFile)
	ml, err := linter.NewMetaLinter(configFile)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	linterType, paths, err := ParseArgs()
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	err = ml.LintFilesAs(linterType, paths)
	if err != nil {
		logrus.Error(err)
		os.Exit(2)
	}
	logrus.Info("Done without errors")
}
